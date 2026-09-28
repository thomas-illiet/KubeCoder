<template>
  <v-card class="auth-card auth-card--compact">
    <div class="auth-card__header">
      <div class="auth-icon"><v-icon icon="mdi-shield-lock-outline" color="warning" size="26" /></div>
      <h1>Access denied</h1>
      <p>Platform administrator access is required to open this area.</p>
    </div>
    <div class="auth-card__body">
      <v-btn block color="primary" prepend-icon="mdi-arrow-left" to="/organization">Return to organization</v-btn>
    </div>
  </v-card>
</template>
