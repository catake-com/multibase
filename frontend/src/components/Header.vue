<script>
import { defineComponent } from "vue";

import { useProjectStore } from "../stores/project";
import { mapState, mapWritableState } from "pinia";

export default defineComponent({
  name: "Header",
  computed: {
    ...mapState(useProjectStore, ["openedProjectIDs", "projects"]),
    ...mapWritableState(useProjectStore, ["currentProjectID"]),
  },
  methods: {
    createNewProject() {
      const store = useProjectStore();

      store.createNewProject();
    },

    closeProjectTab(event, projectID) {
      event.preventDefault();

      const store = useProjectStore();

      store.closeProjectTab(projectID);
    },
  },
});
</script>

<template>
  <div class="bg-primary text-white shadow-2">
    <q-tabs v-model="currentProjectID" align="left" outside-arrows mobile-arrows dense no-caps>
      <q-tab
        :name="parseInt(projectID)"
        v-for="projectID in openedProjectIDs"
        :key="`project-tab-${projectID}`"
        replace
      >
        <div class="row justify-between">
          <div class="col q-tab__label">{{ projects[parseInt(projectID)].type }} {{ projectID }}</div>

          <div class="col-1">
            <q-btn
              class="inline"
              icon="close"
              size="10px"
              style="width: 20px"
              flat
              rounded
              dense
              :disable="openedProjectIDs.length === 1"
              @click="closeProjectTab($event, projectID)"
            />
          </div>
        </div>
      </q-tab>

      <q-btn @click="createNewProject" icon="add" color="primary" />
    </q-tabs>
  </div>
</template>

<style></style>