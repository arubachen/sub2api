<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, useTemplateRef, nextTick } from 'vue'
import alertCircleSvg from '@/assets/icons/alert-circle.svg?raw'

const props = withDefaults(defineProps<{
  content?: string
  trigger?: 'hover' | 'click'
  widthClass?: string
}>(), {
  trigger: 'hover',
  widthClass: 'w-64',
})

const show = ref(false)
const triggerRef = useTemplateRef<HTMLElement>('trigger')
const tooltipRef = useTemplateRef<HTMLElement>('tooltip')
const tooltipStyle = ref({ top: '0px', left: '0px' })
const VIEWPORT_PADDING = 12
const CURSOR_OFFSET = 14
const FOLLOW_EASING = 0.24
const SNAP_THRESHOLD = 0.5

const currentPosition = { top: 0, left: 0 }
const targetPosition = { top: 0, left: 0 }
let frameId: number | null = null

function syncTooltipStyle() {
  tooltipStyle.value = {
    top: `${currentPosition.top}px`,
    left: `${currentPosition.left}px`,
  }
}

function stopAnimation() {
  if (frameId !== null) {
    cancelAnimationFrame(frameId)
    frameId = null
  }
}

function startAnimation() {
  if (frameId !== null) return

  const tick = () => {
    const deltaLeft = targetPosition.left - currentPosition.left
    const deltaTop = targetPosition.top - currentPosition.top

    if (Math.abs(deltaLeft) <= SNAP_THRESHOLD && Math.abs(deltaTop) <= SNAP_THRESHOLD) {
      currentPosition.left = targetPosition.left
      currentPosition.top = targetPosition.top
      syncTooltipStyle()
      frameId = null
      return
    }

    currentPosition.left += deltaLeft * FOLLOW_EASING
    currentPosition.top += deltaTop * FOLLOW_EASING
    syncTooltipStyle()
    frameId = requestAnimationFrame(tick)
  }

  frameId = requestAnimationFrame(tick)
}

function updatePosition(event?: MouseEvent, options?: { immediate?: boolean }) {
  const el = triggerRef.value
  if (!el) return

  const rect = el.getBoundingClientRect()
  const tooltipWidth = tooltipRef.value?.offsetWidth ?? 256
  const tooltipHeight = tooltipRef.value?.offsetHeight ?? 72

  let left = rect.left + rect.width / 2 - tooltipWidth / 2
  let top = rect.top - tooltipHeight - CURSOR_OFFSET

  if (props.trigger === 'hover' && event) {
    left = event.clientX + CURSOR_OFFSET
    top = event.clientY - tooltipHeight - CURSOR_OFFSET

    if (left + tooltipWidth > window.innerWidth - VIEWPORT_PADDING) {
      left = event.clientX - tooltipWidth - CURSOR_OFFSET
    }
    if (top < VIEWPORT_PADDING) {
      top = event.clientY + CURSOR_OFFSET
    }
  }

  left = Math.min(Math.max(left, VIEWPORT_PADDING), window.innerWidth - tooltipWidth - VIEWPORT_PADDING)
  top = Math.min(Math.max(top, VIEWPORT_PADDING), window.innerHeight - tooltipHeight - VIEWPORT_PADDING)

  targetPosition.left = left
  targetPosition.top = top

  if (options?.immediate || props.trigger === 'click') {
    currentPosition.left = left
    currentPosition.top = top
    syncTooltipStyle()
    return
  }

  startAnimation()
}

function openTooltip(event?: MouseEvent) {
  show.value = true
  nextTick(() => {
    updatePosition(event, { immediate: true })
  })
}

function closeTooltip() {
  show.value = false
  stopAnimation()
}

function onEnter(event: MouseEvent) {
  if (props.trigger !== 'hover') return
  openTooltip(event)
}

function onLeave() {
  if (props.trigger !== 'hover') return
  closeTooltip()
}

function onMove(event: MouseEvent) {
  if (props.trigger !== 'hover' || !show.value) return
  updatePosition(event)
}

function onClick(event: MouseEvent) {
  if (props.trigger !== 'click') return
  event.stopPropagation()
  if (show.value) {
    closeTooltip()
    return
  }
  openTooltip()
}

function onDocumentClick(event: MouseEvent) {
  if (props.trigger !== 'click' || !show.value) return
  const target = event.target as Node | null
  if (!target) return
  if (triggerRef.value?.contains(target) || tooltipRef.value?.contains(target)) return
  closeTooltip()
}

function onDocumentKeydown(event: KeyboardEvent) {
  if (props.trigger !== 'click') return
  if (event.key === 'Escape') {
    closeTooltip()
  }
}

function onViewportChange() {
  if (!show.value) return
  updatePosition(undefined, { immediate: true })
}

onMounted(() => {
  document.addEventListener('click', onDocumentClick, true)
  document.addEventListener('keydown', onDocumentKeydown)
  window.addEventListener('resize', onViewportChange)
  window.addEventListener('scroll', onViewportChange, true)
})

onBeforeUnmount(() => {
  stopAnimation()
  document.removeEventListener('click', onDocumentClick, true)
  document.removeEventListener('keydown', onDocumentKeydown)
  window.removeEventListener('resize', onViewportChange)
  window.removeEventListener('scroll', onViewportChange, true)
})
</script>

<template>
  <div
    ref="trigger"
    class="group relative ml-1 inline-flex items-center align-middle"
    @mouseenter="onEnter"
    @mouseleave="onLeave"
    @mousemove="onMove"
    @click="onClick"
  >
    <slot name="trigger">
      <span
        class="inline-flex h-4 w-4 cursor-help text-gray-400 transition-colors hover:text-primary-600 dark:text-gray-500 dark:hover:text-primary-400 [&>svg]:h-4 [&>svg]:w-4"
        v-html="alertCircleSvg"
      ></span>
    </slot>

    <Teleport to="body">
      <Transition
        enter-active-class="transition duration-150 ease-out"
        enter-from-class="translate-y-1 scale-[0.98] opacity-0"
        enter-to-class="translate-y-0 scale-100 opacity-100"
        leave-active-class="transition duration-120 ease-in"
        leave-from-class="translate-y-0 scale-100 opacity-100"
        leave-to-class="translate-y-1 scale-[0.98] opacity-0"
      >
        <div
          v-show="show"
          ref="tooltip"
          role="tooltip"
          :class="[
            'fixed z-[99999] origin-top-left rounded-lg bg-gray-900 p-3 text-xs leading-relaxed text-white shadow-xl ring-1 ring-white/10 will-change-transform dark:bg-gray-800',
            props.widthClass,
            props.trigger === 'click' ? 'pointer-events-auto' : 'pointer-events-none',
          ]"
          :style="tooltipStyle"
        >
          <button
            v-if="props.trigger === 'click'"
            type="button"
            class="absolute right-1.5 top-1.5 rounded p-1 text-gray-300 transition-colors hover:bg-white/10 hover:text-white"
            aria-label="Close"
            @click.stop="closeTooltip"
          >
            <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
          <slot>{{ content }}</slot>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>
