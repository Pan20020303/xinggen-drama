import { computed, ref, type ComputedRef } from "vue";

interface BatchImageGenerationOptions {
  currentEpisode: ComputedRef<any | null>;
  imageDefaultCost: ComputedRef<number>;
  refreshCredits: () => Promise<void>;
  reloadDramaData: () => Promise<void>;
  batchGenerateCharacterImagesRequest: (
    characterIds: string[],
  ) => Promise<unknown>;
  batchGenerateSceneImageRequest: (sceneId: string) => Promise<unknown>;
  notifySuccess: (message: string) => void;
  notifyWarning: (message: string) => void;
  notifyError: (message: string) => void;
  translate: (key: string, params?: Record<string, unknown>) => string;
}

export function useEpisodeBatchImageGeneration({
  currentEpisode,
  imageDefaultCost,
  refreshCredits,
  reloadDramaData,
  batchGenerateCharacterImagesRequest,
  batchGenerateSceneImageRequest,
  notifySuccess,
  notifyWarning,
  notifyError,
  translate,
}: BatchImageGenerationOptions) {
  const selectedCharacterIds = ref<string[]>([]);
  const selectedSceneIds = ref<string[]>([]);
  const selectAllCharacters = ref(false);
  const selectAllScenes = ref(false);
  const batchGeneratingCharacters = ref(false);
  const batchGeneratingScenes = ref(false);

  const batchCharacterImageTotalCost = computed(
    () => imageDefaultCost.value * selectedCharacterIds.value.length,
  );
  const batchSceneImageTotalCost = computed(
    () => imageDefaultCost.value * selectedSceneIds.value.length,
  );

  const toggleSelectAllCharacters = () => {
    if (selectAllCharacters.value) {
      selectedCharacterIds.value =
        currentEpisode.value?.characters?.map((char: any) => char.id) || [];
      return;
    }
    selectedCharacterIds.value = [];
  };

  const toggleSelectAllScenes = () => {
    if (selectAllScenes.value) {
      selectedSceneIds.value =
        currentEpisode.value?.scenes?.map((scene: any) => scene.id) || [];
      return;
    }
    selectedSceneIds.value = [];
  };

  const batchGenerateCharacterImages = async () => {
    if (selectedCharacterIds.value.length === 0) {
      notifyWarning("请先选择要生成的角色");
      return;
    }

    batchGeneratingCharacters.value = true;
    try {
      await batchGenerateCharacterImagesRequest(
        selectedCharacterIds.value.map((id) => id.toString()),
      );
      await refreshCredits();
      notifySuccess(translate("workflow.batchTaskSubmitted"));
      await reloadDramaData();
    } catch (error: any) {
      notifyError(error.message || translate("workflow.batchGenerateFailed"));
    } finally {
      batchGeneratingCharacters.value = false;
    }
  };

  const batchGenerateSceneImages = async () => {
    if (selectedSceneIds.value.length === 0) {
      notifyWarning("请先选择要生成的场景");
      return;
    }

    batchGeneratingScenes.value = true;
    try {
      const results = await Promise.allSettled(
        selectedSceneIds.value.map((sceneId) =>
          batchGenerateSceneImageRequest(sceneId.toString()),
        ),
      );

      const successCount = results.filter(
        (result) => result.status === "fulfilled",
      ).length;
      const failCount = results.filter(
        (result) => result.status === "rejected",
      ).length;

      if (failCount === 0) {
        notifySuccess(
          translate("workflow.batchCompleteSuccess", { count: successCount }),
        );
        return;
      }

      notifyWarning(
        translate("workflow.batchCompletePartial", {
          success: successCount,
          fail: failCount,
        }),
      );
    } catch (error: any) {
      notifyError(error.message || translate("workflow.batchGenerateFailed"));
    } finally {
      batchGeneratingScenes.value = false;
    }
  };

  return {
    selectedCharacterIds,
    selectedSceneIds,
    selectAllCharacters,
    selectAllScenes,
    batchGeneratingCharacters,
    batchGeneratingScenes,
    batchCharacterImageTotalCost,
    batchSceneImageTotalCost,
    toggleSelectAllCharacters,
    toggleSelectAllScenes,
    batchGenerateCharacterImages,
    batchGenerateSceneImages,
  };
}
