import { ref, type ComputedRef, type Ref } from "vue";

interface UseVideoPromptEditorOptions {
  currentStoryboard: ComputedRef<any | null>;
  storyboards: Ref<any[]>;
  textDefaultModel: ComputedRef<string>;
  updateStoryboard: (
    storyboardId: string,
    payload: { video_prompt: string },
  ) => Promise<unknown>;
  optimizeVideoPrompt: (
    storyboardId: string,
    payload: { prompt: string; model?: string },
  ) => Promise<{ prompt?: string }>;
  refreshCredits: () => Promise<void>;
  notifySuccess: (message: string) => void;
  notifyWarning: (message: string) => void;
  notifyError: (message: string) => void;
}

export function useVideoPromptEditor({
  currentStoryboard,
  storyboards,
  textDefaultModel,
  updateStoryboard,
  optimizeVideoPrompt,
  refreshCredits,
  notifySuccess,
  notifyWarning,
  notifyError,
}: UseVideoPromptEditorOptions) {
  const editableVideoPrompt = ref("");
  const optimizingVideoPrompt = ref(false);
  const savingVideoPrompt = ref(false);

  const getDefaultVideoPrompt = (storyboard: any): string => {
    if (!storyboard) return "";
    return (
      storyboard.video_prompt ||
      storyboard.action ||
      storyboard.description ||
      ""
    ).trim();
  };

  const resetVideoPromptEditor = () => {
    editableVideoPrompt.value = getDefaultVideoPrompt(currentStoryboard.value);
  };

  const saveVideoPrompt = async () => {
    if (!currentStoryboard.value) return;

    const prompt = (editableVideoPrompt.value || "").trim();
    if (prompt.length < 5) {
      notifyWarning("提示词至少需要5个字符");
      return;
    }

    savingVideoPrompt.value = true;
    try {
      await updateStoryboard(String(currentStoryboard.value.id), {
        video_prompt: prompt,
      });
      currentStoryboard.value.video_prompt = prompt;
      notifySuccess("视频提示词已保存");
    } catch (error: any) {
      notifyError(error.message || "保存提示词失败");
    } finally {
      savingVideoPrompt.value = false;
    }
  };

  const optimizeVideoPromptWithAI = async () => {
    if (!currentStoryboard.value) return;

    const targetStoryboardId = String(currentStoryboard.value.id);
    const requestPrompt = editableVideoPrompt.value || "";

    optimizingVideoPrompt.value = true;
    try {
      const result = await optimizeVideoPrompt(targetStoryboardId, {
        prompt: requestPrompt,
        model: textDefaultModel.value || undefined,
      });
      const optimized = (result.prompt || "").trim();
      if (!optimized) {
        notifyWarning("未返回可用的优化提示词");
        return;
      }

      const targetStoryboard = storyboards.value.find(
        (storyboard) => String(storyboard.id) === targetStoryboardId,
      );
      if (targetStoryboard) {
        targetStoryboard.video_prompt = optimized;
      }

      if (
        currentStoryboard.value &&
        String(currentStoryboard.value.id) === targetStoryboardId
      ) {
        editableVideoPrompt.value = optimized;
      }

      notifySuccess(
        currentStoryboard.value &&
          String(currentStoryboard.value.id) !== targetStoryboardId
          ? "提示词优化完成，已更新到原镜头"
          : "提示词优化完成",
      );
      await refreshCredits();
    } catch (error: any) {
      notifyError(error.message || "提示词优化失败");
    } finally {
      optimizingVideoPrompt.value = false;
    }
  };

  return {
    editableVideoPrompt,
    optimizingVideoPrompt,
    savingVideoPrompt,
    getDefaultVideoPrompt,
    resetVideoPromptEditor,
    saveVideoPrompt,
    optimizeVideoPromptWithAI,
  };
}
