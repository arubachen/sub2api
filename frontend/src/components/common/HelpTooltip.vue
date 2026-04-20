<script setup lang="ts">
import { ref, useTemplateRef, nextTick } from 'vue'
import alertCircleSvg from '@/assets/icons/alert-circle.svg?raw'

defineProps<{
  content?: string
}>()

const show = ref(false)
const triggerRef = useTemplateRef<HTMLElement>('trigger')
const tooltipRef = useTemplateRef<HTMLElement>('tooltip')
const tooltipStyle = ref({ top: '0px', left: '0px' })
const VIEWPORT_PADDING = 12
const CURSOR_OFFSET = 14

function onEnter(event: MouseEvent) {
  show.value = true
  nextTick(() => updatePosition(event))
}

function onLeave() {
  show.value = false
}

function onMove(event: MouseEvent) {
  if (!show.value) return
  updatePosition(event)
}

function updatePosition(event?: MouseEvent) {
  const el = triggerRef.value
  if (!el) return

  const rect = el.getBoundingClientRect()
  const tooltipWidth = tooltipRef.value?.offsetWidth ?? 256
  const tooltipHeight = tooltipRef.value?.offsetHeight ?? 72

  let left = rect.left + rect.width / 2 - tooltipWidth / 2
  let top = rect.top - tooltipHeight - CURSOR_OFFSET

  if (event) {
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

  tooltipStyle.value = {
    top: `${top}px`,
    left: `${left}px`,
  }
}
</script>

<template>
  <div
    ref="trigger"
    class="group relative ml-1 inline-flex items-center align-middle"
    @mouseenter="onEnter"
    @mouseleave="onLeave"
    @mousemove="onMove"
  >
    <!-- Trigger Icon -->
    <slot name="trigger">
      <span
        class="inline-flex h-4 w-4 cursor-help text-gray-400 transition-colors hover:text-primary-600 dark:text-gray-500 dark:hover:text-primary-400 [&>svg]:h-4 [&>svg]:w-4"
        v-html="alertCircleSvg"
      ></span>
    </slot>

    <!-- Teleport to body to escape modal overflow clipping -->
    <Teleport to="body">
      <div
        ref="tooltip"
        v-show="show"
        class="pointer-events-none fixed z-[99999] w-64 rounded-lg bg-gray-900 p-3 text-xs leading-relaxed text-white shadow-xl ring-1 ring-white/10 dark:bg-gray-800"
        :style="tooltipStyle"
      >
        <slot>{{ content }}</slot>
      </div>
    </Teleport>
  </div>
</template>
