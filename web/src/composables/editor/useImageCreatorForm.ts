import { computed, ref, type ComputedRef, type Ref } from "vue";
import type { ImageGeneration } from "@/types/image";

type ImageCreatorEntityType = "character" | "scene";

interface UseImageCreatorFormOptions {
  imageCreatorItem: Ref<any>;
  imageCreatorType: Ref<ImageCreatorEntityType>;
  imageDefaultModel: ComputedRef<string>;
  imageCreatorVisible: Ref<boolean>;
  imageCreatorHistory: Ref<ImageGeneration[]>;
  imageCreatorSelectedHistoryId: Ref<string | null>;
  imageCreatorPreviewUrl: Ref<string>;
  loadImageCreatorHistory: () => Promise<void>;
  selectHistoryImageAsReference: (
    historyImage: ImageGeneration,
  ) => { imageUrl: string; localPath: string } | null;
  getImageUrl: (item: any) => string;
  notifySuccess: (message: string) => void;
  notifyError: (message: string) => void;
  notifyWarning: (message: string) => void;
}

export function useImageCreatorForm({
  imageCreatorItem,
  imageCreatorType,
  imageDefaultModel,
  imageCreatorVisible,
  imageCreatorHistory,
  imageCreatorSelectedHistoryId,
  imageCreatorPreviewUrl,
  loadImageCreatorHistory,
  selectHistoryImageAsReference,
  getImageUrl,
  notifySuccess,
  notifyError,
  notifyWarning,
}: UseImageCreatorFormOptions) {
  const imageCreatorPrompt = ref("");
  const imageCreatorMode = ref<"text" | "image">("text");
  const imageCreatorModel = ref("");
  const imageCreatorSize = ref("2560x1440");
  const imageCreatorQuality = ref("standard");
  const imageCreatorStyle = ref("vivid");
  const imageCreatorSteps = ref(30);
  const imageCreatorCfgScale = ref(7.5);
  const imageCreatorSeed = ref<number | undefined>(undefined);
  const imageCreatorReferenceUrl = ref("");
  const imageCreatorReferenceLocalPath = ref("");
  const imageCreatorSubmitting = ref(false);

  const imageCreatorTitle = computed(() =>
    imageCreatorType.value === "character" ? "角色图片创作" : "场景图片创作",
  );

  const imageCreatorCurrentImage = computed(() => {
    if (imageCreatorPreviewUrl.value) {
      return imageCreatorPreviewUrl.value;
    }
    const item = imageCreatorItem.value;
    if (!item) return "";
    return getImageUrl(item);
  });

  const imageCreatorCanUseCurrentImage = computed(() =>
    Boolean(imageCreatorItem.value && getImageUrl(imageCreatorItem.value)),
  );

  const imageCreatorSizeOptions = [
    { label: "16:9 横图", value: "2560x1440" },
    { label: "4:3 经典", value: "2048x1536" },
    { label: "1:1 方图", value: "1536x1536" },
    { label: "9:16 竖图", value: "1440x2560" },
  ];

  const getImageCreatorPrompt = (item: any, type: ImageCreatorEntityType) => {
    if (type === "character") {
      return item.appearance || item.description || item.name || "";
    }
    return item.prompt || item.description || item.location || "";
  };

  const formatImageCreatorHistoryTime = (value?: string) => {
    if (!value) return "";
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return value;
    }
    return date.toLocaleString("zh-CN", {
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const resetImageCreatorForm = () => {
    imageCreatorPrompt.value = "";
    imageCreatorModel.value = imageDefaultModel.value || "";
    imageCreatorSize.value = "2560x1440";
    imageCreatorQuality.value = "standard";
    imageCreatorStyle.value = "vivid";
    imageCreatorSteps.value = 30;
    imageCreatorCfgScale.value = 7.5;
    imageCreatorSeed.value = undefined;
    imageCreatorReferenceUrl.value = "";
    imageCreatorReferenceLocalPath.value = "";
    imageCreatorPreviewUrl.value = "";
    imageCreatorSelectedHistoryId.value = null;
    imageCreatorHistory.value = [];
    imageCreatorMode.value = "text";
  };

  const openImageCreator = (item: any, type: ImageCreatorEntityType) => {
    imageCreatorItem.value = item;
    imageCreatorType.value = type;
    resetImageCreatorForm();
    imageCreatorPrompt.value = getImageCreatorPrompt(item, type);
    imageCreatorVisible.value = true;
    void loadImageCreatorHistory();
  };

  const useHistoryImageAsReference = (historyImage: ImageGeneration) => {
    const nextReference = selectHistoryImageAsReference(historyImage);
    if (!nextReference) return;

    imageCreatorReferenceUrl.value = nextReference.imageUrl;
    imageCreatorReferenceLocalPath.value = nextReference.localPath;
    imageCreatorMode.value = "image";
  };

  const handleImageCreatorUploadSuccess = (response: any) => {
    const imageUrl = response.url || response.data?.url;
    const localPath = response.local_path || response.data?.local_path;

    if (!imageUrl || !localPath) {
      notifyError("上传失败：未获取到参考图信息");
      return;
    }

    imageCreatorReferenceUrl.value = imageUrl;
    imageCreatorReferenceLocalPath.value = localPath;
    imageCreatorMode.value = "image";
    notifySuccess("参考图上传成功");
  };

  const useCurrentImageAsReference = () => {
    if (!imageCreatorCanUseCurrentImage.value || !imageCreatorItem.value) {
      notifyWarning("当前没有可作为参考图的图片");
      return;
    }

    imageCreatorReferenceUrl.value = getImageUrl(imageCreatorItem.value);
    imageCreatorReferenceLocalPath.value =
      imageCreatorItem.value.local_path || "";
    imageCreatorMode.value = "image";
    notifySuccess("已将当前图片设为参考图");
  };

  const clearImageCreatorReference = () => {
    imageCreatorReferenceUrl.value = "";
    imageCreatorReferenceLocalPath.value = "";
    if (imageCreatorMode.value === "image") {
      imageCreatorMode.value = "text";
    }
  };

  return {
    imageCreatorPrompt,
    imageCreatorMode,
    imageCreatorModel,
    imageCreatorSize,
    imageCreatorQuality,
    imageCreatorStyle,
    imageCreatorSteps,
    imageCreatorCfgScale,
    imageCreatorSeed,
    imageCreatorReferenceUrl,
    imageCreatorReferenceLocalPath,
    imageCreatorSubmitting,
    imageCreatorTitle,
    imageCreatorCurrentImage,
    imageCreatorCanUseCurrentImage,
    imageCreatorSizeOptions,
    formatImageCreatorHistoryTime,
    handleImageCreatorUploadSuccess,
    useHistoryImageAsReference,
    openImageCreator,
    useCurrentImageAsReference,
    clearImageCreatorReference,
  };
}
