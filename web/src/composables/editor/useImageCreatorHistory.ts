import { computed, ref, type Ref } from "vue";
import type { ImageGeneration } from "@/types/image";

interface UseImageCreatorHistoryOptions {
  dramaId: Ref<string>;
  imageCreatorItem: Ref<any>;
  imageCreatorType: Ref<"character" | "scene">;
  getImageUrl: (item: any) => string;
  listImages: (params: Record<string, any>) => Promise<{ items: ImageGeneration[] }>;
  deleteImage: (id: string) => Promise<void>;
  persistCurrentImage: (nextImage: ImageGeneration | null) => Promise<void>;
  reloadCurrentItem: () => void;
  reloadEpisodeData: () => Promise<void>;
  notifySuccess: (message: string) => void;
  notifyError: (message: string) => void;
  notifyWarning: (message: string) => void;
  confirmDelete: () => Promise<void>;
}

export function useImageCreatorHistory({
  dramaId,
  imageCreatorItem,
  imageCreatorType,
  getImageUrl,
  listImages,
  deleteImage,
  persistCurrentImage,
  reloadCurrentItem,
  reloadEpisodeData,
  notifySuccess,
  notifyError,
  notifyWarning,
  confirmDelete,
}: UseImageCreatorHistoryOptions) {
  const imageCreatorHistory = ref<ImageGeneration[]>([]);
  const imageCreatorHistoryLoading = ref(false);
  const imageCreatorDeletingId = ref<string | null>(null);
  const imageCreatorSelectedHistoryId = ref<string | null>(null);
  const imageCreatorPreviewUrl = ref("");

  const isSameImageResource = (
    left?: { image_url?: string; local_path?: string } | null,
    right?: { image_url?: string; local_path?: string } | null,
  ) => {
    if (!left || !right) return false;
    if (left.local_path && right.local_path) {
      return left.local_path === right.local_path;
    }
    if (left.image_url && right.image_url) {
      return left.image_url === right.image_url;
    }
    return false;
  };

  const imageCreatorSelectedHistory = computed(() =>
    imageCreatorHistory.value.find(
      (item) => String(item.id) === imageCreatorSelectedHistoryId.value,
    ) || null,
  );

  const syncSelectionWithCurrentImage = () => {
    const currentItem = imageCreatorItem.value;
    if (!currentItem) {
      imageCreatorSelectedHistoryId.value = null;
      imageCreatorPreviewUrl.value = "";
      return;
    }

    const matchedHistory = imageCreatorHistory.value.find((historyImage) =>
      isSameImageResource(historyImage, currentItem),
    );

    imageCreatorSelectedHistoryId.value = matchedHistory?.id
      ? String(matchedHistory.id)
      : null;
    imageCreatorPreviewUrl.value = "";
  };

  const loadImageCreatorHistory = async () => {
    if (!imageCreatorItem.value) return;

    imageCreatorHistoryLoading.value = true;
    try {
      const params: Record<string, any> = {
        drama_id: dramaId.value,
        status: "completed",
        page_size: 50,
      };

      if (imageCreatorType.value === "character") {
        params.character_id = String(imageCreatorItem.value.id);
      } else {
        params.scene_id = String(imageCreatorItem.value.id);
      }

      const result = await listImages(params);
      imageCreatorHistory.value = result.items || [];
      syncSelectionWithCurrentImage();
    } catch (error: any) {
      imageCreatorHistory.value = [];
      notifyError(error.message || "加载历史图片失败");
    } finally {
      imageCreatorHistoryLoading.value = false;
    }
  };

  const selectImageCreatorHistoryImage = (historyImage: ImageGeneration) => {
    imageCreatorSelectedHistoryId.value = String(historyImage.id);
    imageCreatorPreviewUrl.value = getImageUrl(historyImage);
  };

  const resetImageCreatorPreview = () => {
    syncSelectionWithCurrentImage();
  };

  const useHistoryImageAsReference = (historyImage: ImageGeneration) => {
    if (!historyImage.local_path) {
      notifyWarning("该历史图片没有本地文件，暂时不能作为图生图参考");
      return null;
    }

    notifySuccess("已设为参考图");
    return {
      imageUrl: getImageUrl(historyImage),
      localPath: historyImage.local_path,
    };
  };

  const deleteImageCreatorHistoryImage = async (historyImage: ImageGeneration) => {
    try {
      await confirmDelete();
    } catch {
      return;
    }

    imageCreatorDeletingId.value = String(historyImage.id);
    try {
      await deleteImage(String(historyImage.id));

      const remainingHistory = imageCreatorHistory.value.filter(
        (item) => item.id !== historyImage.id,
      );
      const deletingCurrentImage =
        !!imageCreatorItem.value &&
        isSameImageResource(historyImage, imageCreatorItem.value);

      imageCreatorHistory.value = remainingHistory;

      if (deletingCurrentImage) {
        await persistCurrentImage(remainingHistory[0] || null);
        await reloadEpisodeData();
        reloadCurrentItem();
      }

      if (imageCreatorSelectedHistoryId.value === String(historyImage.id)) {
        syncSelectionWithCurrentImage();
      }

      notifySuccess("历史图片已删除");
    } catch (error: any) {
      notifyError(error.message || "删除历史图片失败");
    } finally {
      imageCreatorDeletingId.value = null;
    }
  };

  const refreshImageCreatorAfterGeneration = async () => {
    reloadCurrentItem();
    await loadImageCreatorHistory();
    if (imageCreatorHistory.value[0]) {
      selectImageCreatorHistoryImage(imageCreatorHistory.value[0]);
    }
  };

  const isImageCreatorCurrentImage = (historyImage: ImageGeneration) =>
    isSameImageResource(historyImage, imageCreatorItem.value);

  return {
    imageCreatorHistory,
    imageCreatorHistoryLoading,
    imageCreatorDeletingId,
    imageCreatorSelectedHistoryId,
    imageCreatorPreviewUrl,
    imageCreatorSelectedHistory,
    loadImageCreatorHistory,
    selectImageCreatorHistoryImage,
    resetImageCreatorPreview,
    useHistoryImageAsReference,
    deleteImageCreatorHistoryImage,
    refreshImageCreatorAfterGeneration,
    isImageCreatorCurrentImage,
    syncSelectionWithCurrentImage,
  };
}
