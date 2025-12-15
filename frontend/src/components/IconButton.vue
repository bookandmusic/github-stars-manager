<template>
  <button 
    :class="[
      'glass-button p-2 rounded-full w-10 h-10 flex items-center justify-center',
      disabled ? 'opacity-50 cursor-not-allowed' : 'hover:bg-white/20'
    ]"
    :disabled="disabled"
    @click="$emit('click', $event)"
  >
    <slot></slot>
  </button>
</template>

<script setup lang="ts">
interface Props {
  disabled?: boolean
}

withDefaults(defineProps<Props>(), {
  disabled: false
})

defineEmits<{
  (e: 'click', event: MouseEvent): void
}>()
</script>
