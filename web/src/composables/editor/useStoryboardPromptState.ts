import { watch, type ComputedRef, type Ref } from "vue";
import type { FrameType } from "@/api/frame";
import type { ImageGeneration } from "@/types/image";
import type { VideoGeneration } from "@/types/video";

interface UseStoryboardPromptStateOptions {
  currentStoryboard: ComputedRef<any | null>;
  selectedFrameType: Ref<FrameType>;
  currentFramePrompt: Ref<string>;
  framePrompts: Ref<Record<string, string>>;
  generatedImages: Ref<ImageGeneration[]>;
  generatedVideos: Ref<VideoGeneration[]>;
  videoReferenceImages: Ref<ImageGeneration[]>;
  previousStoryboardLastFrames: Ref<any[]>;
  isSwitchingFrameType: Ref<boolean>;
  selectedVideoModel: Ref<string>;
  selectedReferenceMode: Ref<string>;
  selectedImagesForVideo: Ref<string[]>;
  selectedLastImageForVideo: Ref<string | null>;
  videoDuration: Ref<number>;
  editableVideoPrompt: Ref<string>;
  stopPolling: () => void;
  loadStoryboardImages: (storyboardId: string, frameType?: FrameType) => Promise<void>;
  loadAllGeneratedImages: () => Promise<void>;
  loadVideoReferenceImages: (storyboardId: string) => Promise<void>;
  loadStoryboardVideos: (storyboardId: string) => Promise<void>;
  loadPreviousStoryboardLastFrame: () => Promise<void>;
  getDefaultVideoPrompt: (storyboard: any) => string;
}

const emptyFramePrompts = (): Record<string, string> => ({
  key: "",
  first: "",
  last: "",
  panel: "",
});

export function useStoryboardPromptState({
  currentStoryboard,
  selectedFrameType,
  currentFramePrompt,
  framePrompts,
  generatedImages,
  generatedVideos,
  videoReferenceImages,
  previousStoryboardLastFrames,
  isSwitchingFrameType,
  selectedVideoModel,
  selectedReferenceMode,
  selectedImagesForVideo,
  selectedLastImageForVideo,
  videoDuration,
  editableVideoPrompt,
  stopPolling,
  loadStoryboardImages,
  loadAllGeneratedImages,
  loadVideoReferenceImages,
  loadStoryboardVideos,
  loadPreviousStoryboardLastFrame,
  getDefaultVideoPrompt,
}: UseStoryboardPromptStateOptions) {
  const getPromptStorageKey = (
    storyboardId: string | number | undefined,
    frameType: FrameType,
  ) => {
    if (!storyboardId) return null;
    return `frame_prompt_${storyboardId}_${frameType}`;
  };

  watch(selectedFrameType, (newType) => {
    stopPolling();

    if (!currentStoryboard.value) {
      currentFramePrompt.value = "";
      generatedImages.value = [];
      return;
    }

    isSwitchingFrameType.value = true;

    const storageKey = getPromptStorageKey(currentStoryboard.value.id, newType);
    const stored = storageKey ? sessionStorage.getItem(storageKey) : null;

    if (stored) {
      currentFramePrompt.value = stored;
      framePrompts.value[newType] = stored;
    } else {
      currentFramePrompt.value = framePrompts.value[newType] || "";
    }

    void loadStoryboardImages(String(currentStoryboard.value.id), newType);

    setTimeout(() => {
      isSwitchingFrameType.value = false;
    }, 0);
  });

  watch(currentStoryboard, async (newStoryboard) => {
    if (!newStoryboard) {
      currentFramePrompt.value = "";
      generatedImages.value = [];
      generatedVideos.value = [];
      videoReferenceImages.value = [];
      previousStoryboardLastFrames.value = [];
      return;
    }

    isSwitchingFrameType.value = true;
    framePrompts.value = emptyFramePrompts();

    const storageKey = getPromptStorageKey(
      newStoryboard.id,
      selectedFrameType.value,
    );
    const stored = storageKey ? sessionStorage.getItem(storageKey) : null;
    currentFramePrompt.value = stored || "";
    if (stored) {
      framePrompts.value[selectedFrameType.value] = stored;
    }

    setTimeout(() => {
      isSwitchingFrameType.value = false;
    }, 0);

    const storyboardId = String(newStoryboard.id);
    await loadStoryboardImages(storyboardId, selectedFrameType.value);
    await loadAllGeneratedImages();
    await loadVideoReferenceImages(storyboardId);
    await loadStoryboardVideos(storyboardId);
    await loadPreviousStoryboardLastFrame();
  });

  watch(currentFramePrompt, (newPrompt) => {
    if (isSwitchingFrameType.value || !currentStoryboard.value) return;

    const storageKey = getPromptStorageKey(
      currentStoryboard.value.id,
      selectedFrameType.value,
    );
    if (!storageKey) return;

    if (newPrompt) {
      sessionStorage.setItem(storageKey, newPrompt);
      return;
    }

    sessionStorage.removeItem(storageKey);
  });

  watch(selectedVideoModel, () => {
    selectedImagesForVideo.value = [];
    selectedLastImageForVideo.value = null;
    selectedReferenceMode.value = "";
  });

  watch(currentStoryboard, (newStoryboard) => {
    if (newStoryboard?.duration) {
      videoDuration.value = Math.round(newStoryboard.duration);
    } else {
      videoDuration.value = 5;
    }

    editableVideoPrompt.value = getDefaultVideoPrompt(newStoryboard);
  });

  watch(selectedReferenceMode, () => {
    selectedImagesForVideo.value = [];
    selectedLastImageForVideo.value = null;
  });

  return {
    getPromptStorageKey,
  };
}
