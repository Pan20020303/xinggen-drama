import { computed, ref, watch, type ComputedRef, type Ref } from "vue";

export interface ImageReferenceCandidate {
  key: string;
  id: string;
  type: "scene" | "character" | "upload";
  name: string;
  image_url?: string;
  local_path?: string;
}

interface UseImageReferenceSelectorOptions {
  currentStoryboard: ComputedRef<any | null>;
  currentStoryboardCharacters: ComputedRef<any[]>;
  hasImage: (item: any) => boolean;
  maxSelection?: number;
}

export function useImageReferenceSelector({
  currentStoryboard,
  currentStoryboardCharacters,
  hasImage,
  maxSelection = 4,
}: UseImageReferenceSelectorOptions) {
  const selectedImageReferenceKeys = ref<string[]>([]);
  const uploadedImageReferenceMap = ref<Record<string, ImageReferenceCandidate[]>>(
    {},
  );

  const sceneReferenceCandidate = computed<ImageReferenceCandidate | null>(() => {
    const bg = currentStoryboard.value?.background;
    if (!bg || !hasImage(bg)) return null;
    const sceneId = String(
      bg.id || currentStoryboard.value?.scene_id || currentStoryboard.value?.id || "",
    );
    return {
      key: `scene-${sceneId}`,
      id: sceneId,
      type: "scene",
      name: `${bg.location || "场景"}${bg.time ? ` · ${bg.time}` : ""}`,
      image_url: bg.image_url,
      local_path: bg.local_path,
    };
  });

  const characterReferenceCandidates = computed<ImageReferenceCandidate[]>(() => {
    return currentStoryboardCharacters.value
      .filter((char: any) => hasImage(char))
      .map((char: any) => ({
        key: `char-${char.id}`,
        id: String(char.id),
        type: "character" as const,
        name: char.name || `角色${char.id}`,
        image_url: char.image_url,
        local_path: char.local_path,
      }));
  });

  const uploadedReferenceCandidates = computed<ImageReferenceCandidate[]>(() => {
    const storyboardId = String(currentStoryboard.value?.id || "");
    if (!storyboardId) return [];
    return uploadedImageReferenceMap.value[storyboardId] || [];
  });

  const imageReferenceCandidates = computed<ImageReferenceCandidate[]>(() => {
    const result: ImageReferenceCandidate[] = [];
    if (sceneReferenceCandidate.value) {
      result.push(sceneReferenceCandidate.value);
    }
    result.push(...characterReferenceCandidates.value);
    result.push(...uploadedReferenceCandidates.value);
    return result;
  });

  const imageReferenceCandidateMap = computed(() => {
    return new Map(imageReferenceCandidates.value.map((item) => [item.key, item]));
  });

  const selectedImageReferenceItems = computed<ImageReferenceCandidate[]>(() => {
    return selectedImageReferenceKeys.value
      .map((key) => imageReferenceCandidateMap.value.get(key))
      .filter((item): item is ImageReferenceCandidate => !!item);
  });

  const emptyImageReferenceSlotCount = computed(() => {
    return Math.max(0, maxSelection - selectedImageReferenceItems.value.length);
  });

  const selectDefaultImageReferences = () => {
    selectedImageReferenceKeys.value = imageReferenceCandidates.value
      .slice(0, maxSelection)
      .map((item) => item.key);
  };

  const toggleImageReference = (key: string) => {
    const currentIndex = selectedImageReferenceKeys.value.indexOf(key);
    if (currentIndex > -1) {
      selectedImageReferenceKeys.value.splice(currentIndex, 1);
      return { changed: true, reason: "removed" as const };
    }

    if (selectedImageReferenceKeys.value.length >= maxSelection) {
      return { changed: false, reason: "limit" as const };
    }

    selectedImageReferenceKeys.value.push(key);
    return { changed: true, reason: "added" as const };
  };

  const removeImageReference = (key: string) => {
    const index = selectedImageReferenceKeys.value.indexOf(key);
    if (index > -1) {
      selectedImageReferenceKeys.value.splice(index, 1);
    }
  };

  const clearImageReferences = () => {
    selectedImageReferenceKeys.value = [];
  };

  const getImageReferenceTypeLabel = (type: ImageReferenceCandidate["type"]) => {
    if (type === "scene") return "场景";
    if (type === "character") return "人物";
    return "上传";
  };

  const appendUploadedReference = (
    storyboardId: string,
    candidate: ImageReferenceCandidate,
  ) => {
    const uploadedList = uploadedImageReferenceMap.value[storyboardId] || [];
    uploadedImageReferenceMap.value[storyboardId] = [candidate, ...uploadedList];

    if (
      String(currentStoryboard.value?.id || "") === storyboardId &&
      selectedImageReferenceKeys.value.length < maxSelection
    ) {
      selectedImageReferenceKeys.value.push(candidate.key);
    }
  };

  const removeUploadedImageReference = (key: string) => {
    const storyboardId = String(currentStoryboard.value?.id || "");
    if (!storyboardId) return;
    const current = uploadedImageReferenceMap.value[storyboardId] || [];
    uploadedImageReferenceMap.value[storyboardId] = current.filter(
      (item) => item.key !== key,
    );
    removeImageReference(key);
  };

  watch(
    () => currentStoryboard.value?.id,
    (newStoryboardId, oldStoryboardId) => {
      if (!newStoryboardId) {
        selectedImageReferenceKeys.value = [];
        return;
      }
      if (newStoryboardId !== oldStoryboardId) {
        selectDefaultImageReferences();
      }
    },
    { immediate: true },
  );

  watch(
    imageReferenceCandidates,
    (candidates) => {
      const validKeys = new Set(candidates.map((item) => item.key));
      selectedImageReferenceKeys.value = selectedImageReferenceKeys.value.filter(
        (key) => validKeys.has(key),
      );
      if (selectedImageReferenceKeys.value.length === 0 && candidates.length > 0) {
        selectDefaultImageReferences();
      }
    },
    { immediate: true },
  );

  return {
    maxImageReferenceCount: maxSelection,
    selectedImageReferenceKeys,
    uploadedImageReferenceMap,
    sceneReferenceCandidate,
    characterReferenceCandidates,
    uploadedReferenceCandidates,
    imageReferenceCandidates,
    selectedImageReferenceItems,
    emptyImageReferenceSlotCount,
    selectDefaultImageReferences,
    toggleImageReference,
    removeImageReference,
    clearImageReferences,
    getImageReferenceTypeLabel,
    appendUploadedReference,
    removeUploadedImageReference,
  };
}
