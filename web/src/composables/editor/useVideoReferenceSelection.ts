import { computed, ref, type ComputedRef, type Ref } from "vue";

export interface VideoReferenceCandidate {
  id: string;
  image_url?: string;
  local_path?: string;
  frame_type?: string;
  status?: string;
  source_type?: "image" | "character" | "scene";
  name?: string;
}

interface UseVideoReferenceSelectionOptions {
  currentStoryboard: ComputedRef<any | null>;
  previousStoryboard: ComputedRef<any | null>;
  selectedReferenceMode: Ref<string>;
  currentModelCapability: ComputedRef<any>;
  videoReferenceImages: Ref<any[]>;
  hasImage: (item: any) => boolean;
  listPreviousLastFrames: (storyboardId: string) => Promise<any[]>;
  notifyWarning: (message: string) => void;
  notifySuccess?: (message: string) => void;
}

export function useVideoReferenceSelection({
  currentStoryboard,
  previousStoryboard,
  selectedReferenceMode,
  currentModelCapability,
  videoReferenceImages,
  hasImage,
  listPreviousLastFrames,
  notifyWarning,
  notifySuccess,
}: UseVideoReferenceSelectionOptions) {
  const selectedImagesForVideo = ref<string[]>([]);
  const selectedLastImageForVideo = ref<string | null>(null);
  const previousStoryboardLastFrames = ref<any[]>([]);

  const loadPreviousStoryboardLastFrame = async () => {
    if (!previousStoryboard.value?.id) {
      previousStoryboardLastFrames.value = [];
      return;
    }

    try {
      const images = await listPreviousLastFrames(String(previousStoryboard.value.id));
      previousStoryboardLastFrames.value = images.filter(
        (img: any) => img.status === "completed" && hasImage(img),
      );
    } catch (error) {
      console.error("加载上一镜头尾帧失败:", error);
      previousStoryboardLastFrames.value = [];
    }
  };

  const resetSelectedVideoReferences = () => {
    selectedImagesForVideo.value = [];
    selectedLastImageForVideo.value = null;
  };

  const allVideoReferenceCandidates = computed<VideoReferenceCandidate[]>(() => {
    const imageRefs = videoReferenceImages.value
      .filter((img) => img.status === "completed" && hasImage(img))
      .map((img) => ({
        id: String(img.id),
        image_url: img.image_url,
        local_path: img.local_path,
        frame_type: img.frame_type,
        status: img.status,
        source_type: "image" as const,
      }));

    return imageRefs;
  });

  const findReferenceCandidateById = (id: string): VideoReferenceCandidate | any | undefined => {
    return (
      allVideoReferenceCandidates.value.find((img) => img.id === id) ||
      previousStoryboardLastFrames.value.find((img) => String(img.id) === id)
    );
  };

  const handleImageSelect = (imageId: string) => {
    if (!selectedReferenceMode.value) {
      notifyWarning("请先选择参考图模式");
      return;
    }

    if (!currentModelCapability.value) {
      notifyWarning("请先选择视频生成模型");
      return;
    }

    const currentIndex = selectedImagesForVideo.value.indexOf(imageId);
    if (currentIndex > -1) {
      selectedImagesForVideo.value.splice(currentIndex, 1);
      return;
    }

    const clickedImage = findReferenceCandidateById(imageId);
    if (!clickedImage) return;

    switch (selectedReferenceMode.value) {
      case "single":
        selectedImagesForVideo.value = [imageId];
        break;
      case "first_last": {
        const frameType = clickedImage.frame_type;
        if (frameType === "first" || frameType === "panel" || frameType === "key") {
          selectedImagesForVideo.value = [imageId];
        } else if (frameType === "last") {
          selectedLastImageForVideo.value = imageId;
        } else {
          notifyWarning("首尾帧模式下，请选择首帧或尾帧类型的图片");
        }
        break;
      }
      default:
        notifyWarning("未知的参考图模式");
    }
  };

  const selectPreviousLastFrame = (img: any) => {
    const imageId = String(img.id);
    const currentIndex = selectedImagesForVideo.value.indexOf(imageId);
    if (currentIndex > -1) {
      selectedImagesForVideo.value.splice(currentIndex, 1);
      notifySuccess?.("已取消首帧参考");
      return;
    }

    if (!selectedReferenceMode.value || selectedReferenceMode.value === "single") {
      selectedImagesForVideo.value = [imageId];
    } else if (selectedReferenceMode.value === "first_last") {
      selectedImagesForVideo.value = [imageId];
    } else {
      notifyWarning("当前模式不支持该操作");
      return;
    }
    notifySuccess?.("已添加为首帧参考");
  };

  const selectedImageObjects = computed(() => {
    return selectedImagesForVideo.value
      .map((id) => findReferenceCandidateById(id))
      .filter((img) => img && hasImage(img));
  });

  const firstFrameSlotImage = computed(() => {
    if (selectedImagesForVideo.value.length === 0) return null;
    return findReferenceCandidateById(selectedImagesForVideo.value[0]) || null;
  });

  const lastFrameSlotImage = computed(() => {
    if (!selectedLastImageForVideo.value) return null;
    return findReferenceCandidateById(selectedLastImageForVideo.value) || null;
  });

  const removeSelectedImage = (imageId: string) => {
    if (selectedLastImageForVideo.value === imageId) {
      selectedLastImageForVideo.value = null;
      return;
    }

    const index = selectedImagesForVideo.value.indexOf(imageId);
    if (index > -1) {
      selectedImagesForVideo.value.splice(index, 1);
    }
  };

  return {
    selectedImagesForVideo,
    selectedLastImageForVideo,
    previousStoryboardLastFrames,
    allVideoReferenceCandidates,
    selectedImageObjects,
    firstFrameSlotImage,
    lastFrameSlotImage,
    loadPreviousStoryboardLastFrame,
    resetSelectedVideoReferences,
    findReferenceCandidateById,
    handleImageSelect,
    removeSelectedImage,
    selectPreviousLastFrame,
  };
}
