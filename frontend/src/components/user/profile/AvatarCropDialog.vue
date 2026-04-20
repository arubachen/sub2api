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
            class="relative h-60 w-60 overflow-hidden rounded-[2rem] border border-gray-200 bg-gray-100 dark:border-dark-700 dark:bg-dark-900"
            @mousedown="startDrag"
            @touchstart.prevent="startTouchDrag"
            @wheel.prevent="handleWheelZoom"
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
          <div class="space-y-3">
            <div class="flex items-center justify-between gap-3">
              <label class="input-label !mb-0">{{ t('profile.avatarZoom') }}</label>
              <span class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ zoomPercent }}</span>
            </div>
            <div class="flex items-center gap-3">
              <button type="button" class="btn btn-secondary btn-sm !h-9 !w-9 !rounded-full !px-0" @click="nudgeScale(-0.1)">
                −
              </button>
              <input
                v-model.number="scale"
                type="range"
                :min="MIN_SCALE"
                :max="MAX_SCALE"
                step="0.01"
                class="avatar-zoom-range"
              />
              <button type="button" class="btn btn-secondary btn-sm !h-9 !w-9 !rounded-full !px-0" @click="nudgeScale(0.1)">
                +
              </button>
            </div>
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
const MIN_SCALE = 1
const MAX_SCALE = 3

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
const naturalWidth = ref(1)
const naturalHeight = ref(1)
const scale = ref(MIN_SCALE)
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
const zoomPercent = computed(() => `${Math.round(scale.value * 100)}%`)

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
  scale.value = MIN_SCALE
  offsetX.value = 0
  offsetY.value = 0
}

function setScale(nextScale: number) {
  scale.value = Math.max(MIN_SCALE, Math.min(MAX_SCALE, Number(nextScale.toFixed(2))))
}

function nudgeScale(delta: number) {
  setScale(scale.value + delta)
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

function handleWheelZoom(event: WheelEvent) {
  const direction = event.deltaY < 0 ? 0.08 : -0.08
  nudgeScale(direction)
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

<style scoped>
.avatar-zoom-range {
  width: 100%;
  height: 0.5rem;
  cursor: pointer;
  appearance: none;
  border-radius: 9999px;
  background: linear-gradient(90deg, rgba(37, 99, 235, 0.9), rgba(124, 58, 237, 0.92));
}

.avatar-zoom-range::-webkit-slider-thumb {
  width: 1rem;
  height: 1rem;
  border: 2px solid #ffffff;
  border-radius: 9999px;
  background: #1d4ed8;
  box-shadow: 0 8px 18px -10px rgba(29, 78, 216, 0.9);
  cursor: pointer;
  appearance: none;
}

.avatar-zoom-range::-moz-range-thumb {
  width: 1rem;
  height: 1rem;
  border: 2px solid #ffffff;
  border-radius: 9999px;
  background: #1d4ed8;
  box-shadow: 0 8px 18px -10px rgba(29, 78, 216, 0.9);
  cursor: pointer;
}

.dark .avatar-zoom-range {
  background: linear-gradient(90deg, rgba(96, 165, 250, 0.92), rgba(167, 139, 250, 0.96));
}

.dark .avatar-zoom-range::-webkit-slider-thumb,
.dark .avatar-zoom-range::-moz-range-thumb {
  background: #dbeafe;
  border-color: rgba(15, 23, 42, 0.9);
  box-shadow: 0 8px 18px -10px rgba(96, 165, 250, 0.85);
}
</style>
