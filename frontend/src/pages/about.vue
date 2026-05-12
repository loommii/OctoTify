<script setup lang="ts">
useHead({
  title: 'About - OctoTify',
  meta: [
    {
      name: 'description',
      content: 'The about page for OctoTify!'
    }
  ]
})
</script>

<template>
  <div class="about">
    <h1>This is an about page</h1>
  </div>
</template>

<style lang="scss" scoped>
.about {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
