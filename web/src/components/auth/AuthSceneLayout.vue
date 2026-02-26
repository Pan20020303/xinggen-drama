<template>
  <div class="auth-scene">
    <div class="auth-bg" />
    <div class="auth-overlay" />

    <div class="brand">xinggen</div>

    <div class="panel-wrap">
      <section class="auth-panel">
        <header class="panel-header">
          <p v-if="kicker" class="kicker">{{ kicker }}</p>
          <h1>{{ title }}</h1>
          <p class="subtitle">{{ subtitle }}</p>
        </header>

        <div class="panel-body">
          <slot />
        </div>

        <footer v-if="$slots.footer" class="panel-footer">
          <slot name="footer" />
        </footer>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  subtitle: string
  kicker?: string
}>()
</script>

<style scoped>
.auth-scene {
  position: relative;
  min-height: 100vh;
  overflow: hidden;
}

.auth-bg {
  position: absolute;
  inset: 0;
  background:
    linear-gradient(120deg, rgba(3, 24, 43, 0.38), rgba(2, 44, 73, 0.2)),
    url('/auth/ocean-login-bg.png') center center / cover no-repeat;
  transform: scale(1.02);
}

.auth-overlay {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(circle at 78% 42%, rgba(5, 11, 17, 0.14), rgba(5, 11, 17, 0.46)),
    linear-gradient(90deg, rgba(5, 10, 16, 0.08), rgba(5, 10, 16, 0.46));
}

.brand {
  position: relative;
  z-index: 2;
  padding: 42px 0 0 56px;
  font-size: 24px;
  line-height: 1;
  font-weight: 700;
  letter-spacing: 0.02em;
  color: rgba(255, 255, 255, 0.9);
  text-shadow: 0 2px 12px rgba(0, 0, 0, 0.2);
}

.panel-wrap {
  position: relative;
  z-index: 2;
  min-height: calc(100vh - 84px);
  padding: 40px min(8vw, 120px) 40px 20px;
  display: flex;
  justify-content: flex-end;
  align-items: center;
}

.auth-panel {
  width: min(420px, 100%);
  border-radius: 18px;
  border: 1px solid rgba(255, 255, 255, 0.11);
  background: rgba(9, 18, 28, 0.72);
  backdrop-filter: blur(14px);
  box-shadow: 0 30px 80px rgba(0, 0, 0, 0.38);
  color: #e9f2ff;
  padding: 34px 30px 24px;
}

.panel-header {
  margin-bottom: 22px;
}

.kicker {
  margin: 0 0 8px;
  font-size: 11px;
  letter-spacing: 0.16em;
  color: rgba(173, 208, 245, 0.82);
}

.panel-header h1 {
  margin: 0;
  font-size: 28px;
  line-height: 1.25;
  color: #fff;
  letter-spacing: 0.01em;
}

.subtitle {
  margin: 8px 0 0;
  font-size: 13px;
  color: rgba(219, 233, 255, 0.66);
}

.panel-footer {
  margin-top: 16px;
}

@media (max-width: 1024px) {
  .brand {
    padding-left: 24px;
    font-size: 21px;
  }

  .panel-wrap {
    justify-content: center;
    padding: 26px 18px 30px;
    min-height: calc(100vh - 78px);
  }

  .auth-panel {
    width: min(460px, 100%);
    padding: 28px 22px 20px;
  }
}
</style>
