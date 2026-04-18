<template>
  <div
    :class="[
      'flex items-center justify-center overflow-hidden bg-gradient-primary text-white',
      roundedClass,
      sizeClass,
      shadowClass,
    ]"
  >
    <img
      v-if="avatarUrl"
      :src="avatarUrl"
      alt=""
      class="h-full w-full object-cover"
    />
    <span v-else :class="textClass">{{ initials }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { sanitizeUrl } from '@/utils/url'

const props = withDefaults(defineProps<{
  avatarUrl?: string | null
  username?: string | null
  email?: string | null
  sizeClass?: string
  textClass?: string
  roundedClass?: string
  shadowClass?: string
}>(), {
  avatarUrl: '',
  username: '',
  email: '',
  sizeClass: 'h-10 w-10',
  textClass: 'text-sm font-semibold',
  roundedClass: 'rounded-full',
  shadowClass: '',
})

const avatarUrl = computed(() => sanitizeUrl(props.avatarUrl || '', { allowDataUrl: true, allowRelative: false }))

const initials = computed(() => {
  if (props.username?.trim()) {
    return props.username.trim().slice(0, 2).toUpperCase()
  }
  if (props.email?.trim()) {
    return props.email.trim().slice(0, 2).toUpperCase()
  }
  return 'U'
})
</script>
