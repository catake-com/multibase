<script>
import Header from "./components/Header.vue";
import Project from "./components/Project.vue";
import { useQuasar } from "quasar";

export default {
  components: { Header, Project },
  setup() {
    const $q = useQuasar();

    $q.dark.set(true);
  },
};
</script>

<template>
  <q-layout view="hHh lpR fFf">
    <Header />

    <q-page-container>
      <Project class="window-height" />
    </q-page-container>
  </q-layout>
</template>

<style></style>
