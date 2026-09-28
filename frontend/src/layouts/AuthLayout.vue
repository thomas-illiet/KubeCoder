<template>
  <v-app class="auth-layout">
    <v-main>
      <div class="auth-shell">
        <router-link to="/login" class="auth-brand" aria-label="KubeCoder login">
          <div class="brand__mark"><v-icon icon="mdi-kubernetes" size="26" /></div>
          <div class="brand__copy"><span class="brand__name">KubeCoder</span><span class="brand__tag">DEVELOPER WORKSPACE</span></div>
        </router-link>
        <router-view />
        <div class="auth-footer">Secure access · Mock interface</div>
      </div>
    </v-main>
  </v-app>
</template>
