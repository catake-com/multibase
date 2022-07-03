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
  beforeCreate() {
    useProjectStore().loadState();
  },
  watch: {
    async currentProjectID(newValue, oldValue) {
      if (newValue === oldValue) {
        return;
      }

      const projectID = newValue || oldValue;

      await useProjectStore().saveCurrentProjectID(projectID);
    },
  },
  methods: {
    async createNewProject() {
      const store = useProjectStore();

      await store.createNewProject();
    },

    async closeProjectTab(event, projectID) {
      event.preventDefault();

      const store = useProjectStore();

      await store.closeProjectTab(projectID);
    },
  },
});
</script>

<template>
  <div class="bg-primary text-white shadow-2">
    <q-tabs v-model="currentProjectID" align="left" outside-arrows mobile-arrows no-caps>
      <q-tab :name="projectID" v-for="projectID in openedProjectIDs" :key="`project-tab-${projectID}`">
        <div class="row justify-between">
          <div class="col q-tab__label" v-if="projects[projectID].type === 'new'">New Tab</div>
          <div class="col q-tab__label" v-if="projects[projectID].type === 'grpc'">{{ projects[projectID].name }}</div>
          <div class="col q-tab__label" v-if="projects[projectID].type === 'thrift'">
            {{ projects[projectID].name }}
          </div>

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
