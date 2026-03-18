import { computed, onUnmounted, ref, type ComputedRef, type Ref } from "vue";
import type { AsyncTask } from "@/api/task";

interface UseStoryboardTaskTrackerOptions {
  episodeId: ComputedRef<string | null>;
  currentEpisode: ComputedRef<any | null>;
  generatingShots: Ref<boolean>;
  currentStep: Ref<number>;
  loadTaskStatus: (taskId: string) => Promise<AsyncTask>;
  listTasksByResource: (resourceId: string) => Promise<AsyncTask[]>;
  reloadDramaData: () => Promise<void>;
  refreshCredits: () => Promise<void>;
  notifySuccess: (message: string) => void;
  notifyError: (message: string) => void;
  taskCreatedMessage?: string;
  taskResumingMessage?: string;
  splitSuccessMessage: string;
}

export function useStoryboardTaskTracker({
  episodeId,
  currentEpisode,
  generatingShots,
  currentStep,
  loadTaskStatus,
  listTasksByResource,
  reloadDramaData,
  refreshCredits,
  notifySuccess,
  notifyError,
  taskCreatedMessage = "任务已创建",
  taskResumingMessage = "正在恢复分镜拆分任务...",
  splitSuccessMessage,
}: UseStoryboardTaskTrackerOptions) {
  const taskProgress = ref(0);
  const taskMessage = ref("");
  let pollTimer: ReturnType<typeof setInterval> | null = null;

  const getStoryboardTaskStorageKey = (currentEpisodeId: string) =>
    `episode_workflow_storyboard_task_${currentEpisodeId}`;

  const setActiveStoryboardTaskId = (currentEpisodeId: string, taskId: string) => {
    localStorage.setItem(getStoryboardTaskStorageKey(currentEpisodeId), taskId);
  };

  const getActiveStoryboardTaskId = (currentEpisodeId: string) =>
    localStorage.getItem(getStoryboardTaskStorageKey(currentEpisodeId));

  const clearActiveStoryboardTaskId = (currentEpisodeId?: string | null) => {
    if (!currentEpisodeId) return;
    localStorage.removeItem(getStoryboardTaskStorageKey(currentEpisodeId));
  };

  const parseTaskResult = (result: any) => {
    if (!result) return null;
    if (typeof result === "string") {
      try {
        return JSON.parse(result);
      } catch (error) {
        console.error("[任务] 解析任务结果失败:", error);
        return null;
      }
    }
    return result;
  };

  const normalizePreviewStoryboards = (storyboards: any[] = []) => {
    const characterMap = new Map(
      (currentEpisode.value?.characters || []).map((char: any) => [
        Number(char.id),
        char,
      ]),
    );

    return storyboards.map((storyboard, index) => {
      const shotNumber =
        Number(
          storyboard.storyboard_number ??
            storyboard.shot_number ??
            storyboard.shotNumber ??
            index + 1,
        ) || index + 1;

      const normalizedCharacters = Array.isArray(storyboard.characters)
        ? storyboard.characters.map((character: any) => {
            if (
              character &&
              typeof character === "object" &&
              ("name" in character || "id" in character)
            ) {
              return character;
            }
            const characterId = Number(character);
            return characterMap.get(characterId) || character;
          })
        : [];

      return {
        ...storyboard,
        storyboard_number: shotNumber,
        shot_number: shotNumber,
        characters: normalizedCharacters,
      };
    });
  };

  const applyStoryboardTaskPreview = (task: { result?: any }) => {
    const parsed = parseTaskResult(task.result);
    if (!parsed || !Array.isArray(parsed.storyboards) || !currentEpisode.value) {
      return;
    }
    currentEpisode.value.storyboards = normalizePreviewStoryboards(parsed.storyboards);
  };

  const stopPollingStoryboardTask = () => {
    if (!pollTimer) return;
    clearInterval(pollTimer);
    pollTimer = null;
  };

  const findActiveStoryboardTask = async (
    currentEpisodeId: string,
  ): Promise<AsyncTask | null> => {
    const storedTaskId = getActiveStoryboardTaskId(currentEpisodeId);

    if (storedTaskId) {
      try {
        const storedTask = await loadTaskStatus(storedTaskId);
        if (
          storedTask.type === "storyboard_generation" &&
          (storedTask.status === "pending" || storedTask.status === "processing")
        ) {
          return storedTask;
        }
      } catch (error) {
        console.warn("[任务] 读取缓存分镜任务失败:", error);
      }
      clearActiveStoryboardTaskId(currentEpisodeId);
    }

    const tasks = await listTasksByResource(currentEpisodeId);
    const activeTask = tasks.find(
      (task) =>
        task.type === "storyboard_generation" &&
        (task.status === "pending" || task.status === "processing"),
    );

    if (activeTask) {
      setActiveStoryboardTaskId(currentEpisodeId, activeTask.id);
      return activeTask;
    }

    return null;
  };

  const pollStoryboardTaskStatus = async (taskId: string) => {
    stopPollingStoryboardTask();

    const checkStatus = async () => {
      try {
        const task = await loadTaskStatus(taskId);

        taskProgress.value = task.progress;
        taskMessage.value = task.message || `处理中... ${task.progress}%`;
        applyStoryboardTaskPreview(task);

        if (task.status === "completed") {
          stopPollingStoryboardTask();
          generatingShots.value = false;
          clearActiveStoryboardTaskId(episodeId.value);
          await refreshCredits();
          await reloadDramaData();
          notifySuccess(splitSuccessMessage);
          return true;
        }

        if (task.status === "failed") {
          stopPollingStoryboardTask();
          generatingShots.value = false;
          clearActiveStoryboardTaskId(episodeId.value);
          await refreshCredits();
          notifyError(task.error || "分镜拆分失败");
          return true;
        }

        return false;
      } catch (error: any) {
        stopPollingStoryboardTask();
        generatingShots.value = false;
        notifyError("查询任务状态失败: " + error.message);
        return true;
      }
    };

    const finished = await checkStatus();
    if (!finished) {
      pollTimer = setInterval(() => {
        void checkStatus();
      }, 2000);
    }
  };

  const resumeStoryboardTaskIfNeeded = async () => {
    const currentEpisodeId = episodeId.value;
    if (!currentEpisodeId || generatingShots.value || pollTimer) return;

    try {
      const activeTask = await findActiveStoryboardTask(currentEpisodeId);
      if (!activeTask) return;

      generatingShots.value = true;
      currentStep.value = 2;
      taskProgress.value = activeTask.progress || 0;
      taskMessage.value = activeTask.message || taskResumingMessage;
      applyStoryboardTaskPreview(activeTask);
      await pollStoryboardTaskStatus(activeTask.id);
    } catch (error) {
      console.error("[任务] 恢复分镜任务失败:", error);
    }
  };

  const startStoryboardTaskTracking = async (
    taskId: string,
    initialMessage?: string,
  ) => {
    const currentEpisodeId = episodeId.value;
    if (!currentEpisodeId) return;

    setActiveStoryboardTaskId(currentEpisodeId, taskId);
    taskMessage.value = initialMessage || taskCreatedMessage;
    await pollStoryboardTaskStatus(taskId);
  };

  const taskPhaseLabel = computed(() => {
    if (!generatingShots.value) return "";
    if (taskProgress.value < 10) return "准备任务";
    if (taskProgress.value < 55) return "AI 拆分中";
    if (taskProgress.value < 70) return "合并结果";
    if (taskProgress.value < 90) return "保存分镜";
    if (taskProgress.value < 100) return "更新章节";
    return "已完成";
  });

  const taskPhaseHint = computed(() => {
    if (!generatingShots.value) return "";
    if (taskProgress.value < 55) {
      return "长章节会拆成多段并行生成，已完成的镜头会逐步显示在下方。";
    }
    if (taskProgress.value < 90) {
      return "已生成的镜头会先展示，系统正在继续保存和整理。";
    }
    return "正在完成最后的整理工作。";
  });

  onUnmounted(() => {
    stopPollingStoryboardTask();
  });

  return {
    taskProgress,
    taskMessage,
    taskPhaseLabel,
    taskPhaseHint,
    applyStoryboardTaskPreview,
    resumeStoryboardTaskIfNeeded,
    startStoryboardTaskTracking,
    pollStoryboardTaskStatus,
    stopPollingStoryboardTask,
  };
}
