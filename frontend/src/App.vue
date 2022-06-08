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
  <q-layout view="hHh lpR fFf" class="window-height">
    <Header />

    <q-page-container id="pageContainer">
      <Project class="full-height" />
    </q-page-container>
  </q-layout>
</template>

<style>
#pageContainer {
  height: calc(100% - 48px) !important;
}
</style>
