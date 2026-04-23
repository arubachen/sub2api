<template>
  <svg
    viewBox="0 0 100 100"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    aria-hidden="true"
  >
    <defs>
      <linearGradient :id="gradientId" x1="10" y1="18" x2="90" y2="82" gradientUnits="userSpaceOnUse">
        <stop offset="0%" stop-color="#22d3ee" />
        <stop offset="55%" stop-color="#06b6d4" />
        <stop offset="100%" stop-color="#34d399" />
      </linearGradient>
    </defs>

    <path
      d="M10 20C30 20 40 40 50 50"
      :stroke="`url(#${gradientId})`"
      stroke-width="8"
      stroke-linecap="round"
    />
    <path
      d="M10 50C30 50 40 50 50 50"
      :stroke="`url(#${gradientId})`"
      stroke-width="8"
      stroke-linecap="round"
    />
    <path
      d="M10 80C30 80 40 60 50 50"
      :stroke="`url(#${gradientId})`"
      stroke-width="8"
      stroke-linecap="round"
    />

    <circle cx="50" cy="50" r="14" fill="#071120" stroke="#67e8f9" stroke-width="2.5" />
    <circle cx="50" cy="50" r="8.5" :fill="`url(#${gradientId})`" />

    <path d="M50 50H90" stroke="#38bdf8" stroke-width="9" stroke-linecap="round" />
    <path
      d="M79 39L90 50L79 61"
      stroke="#67e8f9"
      stroke-width="7"
      stroke-linecap="round"
      stroke-linejoin="round"
    />
  </svg>
</template>

<script setup lang="ts">
const gradientId = `juliu-flow-gradient-${Math.random().toString(36).slice(2, 9)}`
</script>
