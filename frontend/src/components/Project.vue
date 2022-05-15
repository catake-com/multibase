<script>
import { defineComponent } from "vue";
import { mapState } from "pinia";

import { useProjectStore } from "../stores/project";
import GRPC from "./GRPC.vue";

export default defineComponent({
  name: "Project",
  components: { GRPC },
  computed: {
    ...mapState(useProjectStore, ["projects", "currentProjectID"]),
  },
  beforeCreate() {
    useProjectStore().loadState();
  },
});
</script>

<template>
  <q-tab-panels v-model="currentProjectID">
    <q-tab-panel
      :name="parseInt(projectID)"
      v-for="(project, projectID) in projects"
      :key="`project-panel-${projectID}`"
    >
      <div v-if="project.type === 'new'">new</div>

      <div v-if="project.type === 'grpc'">
        <GRPC :projectID="parseInt(projectID)" />
      </div>
    </q-tab-panel>
  </q-tab-panels>
</template>

<style></style>
