<template>
  <BaseDialog
    :show="show"
    :title="t('profile.avatarCropTitle')"
    width="wide"
    close-on-click-outside
    @close="emit('close')"
  >
    <div class="space-y-6">
      <div class="flex flex-col gap-6 lg:flex-row">
        <div class="flex flex-1 items-center justify-center">
          <div
            ref="previewRef"
            class="relative h-60 w-60 overflow-hidden rounded-[2rem] border border-gray-200 bg-gray-100 dark:border-dark-700 dark:bg-dark-900"
            @mousedown="startDrag"
            @touchstart.prevent="startTouchDrag"
          >
            <img
              v-if="imageUrl"
              ref="imageRef"
              :src="imageUrl"
              alt=""
              draggable="false"
              :style="imageStyle"
              class="absolute left-1/2 top-1/2 select-none"
              @load="handleImageLoad"
            />
            <div class="pointer-events-none absolute inset-0 rounded-[2rem] ring-1 ring-inset ring-white/70 dark:ring-white/10"></div>
          </div>
        </div>

        <div class="w-full space-y-4 lg:max-w-xs">
          <div>
            <label class="input-label">{{ t('profile.avatarZoom') }}</label>
            <input
              v-model.number="scale"
              type="range"
              min="1"
              max="3"
              step="0.01"
              class="w-full"
            />
          </div>
          <p class="text-xs text-gray-500 dark:text-gray-400">
            {{ t('profile.avatarCropHint') }}
          </p>
          <div class="flex flex-wrap gap-2">
            <button type="button" class="btn btn-secondary btn-sm" @click="resetCrop">
              {{ t('profile.avatarReset') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="flex w-full justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="emit('close')">
          {{ t('common.cancel') }}
        </button>
        <button type="button" class="btn btn-primary" :disabled="loading" @click="applyCrop">
          {{ loading ? t('profile.updating') : t('profile.avatarApply') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { cropAvatarToDataUrl } from '@/utils/avatar'

const PREVIEW_SIZE = 240

defineProps<{
  show: boolean
  imageUrl: string
}>()

const emit = defineEmits<{
  close: []
  apply: [value: string]
}>()

const { t } = useI18n()

const imageRef = ref<HTMLImageElement | null>(null)
const previewRef = ref<HTMLDivElement | null>(null)
const naturalWidth = ref(1)
const naturalHeight = ref(1)
const scale = ref(1)
const offsetX = ref(0)
const offsetY = ref(0)
const loading = ref(false)

const dragState = {
  active: false,
  startX: 0,
  startY: 0,
  startOffsetX: 0,
  startOffsetY: 0,
}

const renderedSize = computed(() => {
  const ratio = naturalWidth.value / naturalHeight.value
  const baseWidth = ratio >= 1 ? PREVIEW_SIZE * ratio : PREVIEW_SIZE
  const baseHeight = ratio >= 1 ? PREVIEW_SIZE : PREVIEW_SIZE / Math.max(ratio, 0.0001)
  return {
    width: baseWidth * scale.value,
    height: baseHeight * scale.value,
  }
})

const maxOffsets = computed(() => ({
  x: Math.max(0, (renderedSize.value.width - PREVIEW_SIZE) / 2),
  y: Math.max(0, (renderedSize.value.height - PREVIEW_SIZE) / 2),
}))

const imageStyle = computed(() => ({
  width: `${renderedSize.value.width}px`,
  height: `${renderedSize.value.height}px`,
  transform: `translate(calc(-50% + ${offsetX.value}px), calc(-50% + ${offsetY.value}px))`,
  cursor: dragState.active ? 'grabbing' : 'grab',
}))

watch(scale, () => {
  clampOffsets()
})

function clampOffsets() {
  offsetX.value = Math.max(-maxOffsets.value.x, Math.min(maxOffsets.value.x, offsetX.value))
  offsetY.value = Math.max(-maxOffsets.value.y, Math.min(maxOffsets.value.y, offsetY.value))
}

function handleImageLoad() {
  if (!imageRef.value) return
  naturalWidth.value = imageRef.value.naturalWidth
  naturalHeight.value = imageRef.value.naturalHeight
  resetCrop()
}

function resetCrop() {
  scale.value = 1
  offsetX.value = 0
  offsetY.value = 0
}

function startDrag(event: MouseEvent) {
  dragState.active = true
  dragState.startX = event.clientX
  dragState.startY = event.clientY
  dragState.startOffsetX = offsetX.value
  dragState.startOffsetY = offsetY.value
  window.addEventListener('mousemove', handleDrag)
  window.addEventListener('mouseup', stopDrag)
}

function startTouchDrag(event: TouchEvent) {
  const touch = event.touches[0]
  if (!touch) return
  dragState.active = true
  dragState.startX = touch.clientX
  dragState.startY = touch.clientY
  dragState.startOffsetX = offsetX.value
  dragState.startOffsetY = offsetY.value
  window.addEventListener('touchmove', handleTouchDrag, { passive: false })
  window.addEventListener('touchend', stopTouchDrag)
}

function handleDrag(event: MouseEvent) {
  offsetX.value = dragState.startOffsetX + event.clientX - dragState.startX
  offsetY.value = dragState.startOffsetY + event.clientY - dragState.startY
  clampOffsets()
}

function handleTouchDrag(event: TouchEvent) {
  const touch = event.touches[0]
  if (!touch) return
  offsetX.value = dragState.startOffsetX + touch.clientX - dragState.startX
  offsetY.value = dragState.startOffsetY + touch.clientY - dragState.startY
  clampOffsets()
}

function stopDrag() {
  dragState.active = false
  window.removeEventListener('mousemove', handleDrag)
  window.removeEventListener('mouseup', stopDrag)
}

function stopTouchDrag() {
  dragState.active = false
  window.removeEventListener('touchmove', handleTouchDrag)
  window.removeEventListener('touchend', stopTouchDrag)
}

async function applyCrop() {
  if (!imageRef.value) return
  loading.value = true
  try {
    const result = await cropAvatarToDataUrl({
      image: imageRef.value,
      scale: scale.value,
      offsetX: offsetX.value,
      offsetY: offsetY.value,
      previewSize: PREVIEW_SIZE,
    })
    emit('apply', result)
  } finally {
    loading.value = false
  }
}

onBeforeUnmount(() => {
  stopDrag()
  stopTouchDrag()
})
</script>
