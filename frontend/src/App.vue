<script setup>
import { RouterView } from "vue-router";
</script>

<template>
  <q-layout view="hHh lpR fFf">
    <div class="bg-primary text-white shadow-2">
      <q-tabs align="left">
        <q-route-tab to="/" label="Home" />
        <q-route-tab to="/grpc" label="gRPC" />
        <q-route-tab to="/thrift" label="Thrift" />
      </q-tabs>
    </div>

    <q-page-container>
      <router-view />
    </q-page-container>
  </q-layout>
</template>

<style></style>
