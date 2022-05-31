<script>
import { defineComponent } from "vue";
import { mapState } from "pinia";

import { useProjectStore } from "../stores/project";
import { useGRPCStore } from "../stores/grpc";
import GRPC from "./GRPC.vue";

export default defineComponent({
  name: "Project",
  components: { GRPC },
  computed: {
    ...mapState(useProjectStore, ["projects", "currentProjectID"]),
    ...mapState(useGRPCStore, { grpcProjects: "projects" }),
    nonNewProjects() {
      return Object.fromEntries(
        Object.entries(useProjectStore().projects).filter(([projectID, project]) => project.type !== "new")
      );
    },
  },
  beforeCreate() {
    useProjectStore().loadState();
  },
  methods: {
    openGRPCProject(newProjectID, grpcProjectID) {
      const store = useProjectStore();

      store.openGRPCProject(newProjectID, grpcProjectID);
    },

    newGRPCProject() {
      useGRPCStore().createNewProject(this.currentProjectID);
      useProjectStore().createNewGRPCProject(this.currentProjectID);
    },
  },
});
</script>

<template>
  <q-tab-panels v-model="currentProjectID">
    <q-tab-panel :name="projectID" v-for="(project, projectID) in projects" :key="`project-panel-${projectID}`">
      <div v-if="project.type === 'new'">
        <div class="row q-col-gutter-sm" style="margin-top: 10%">
          <div class="col"></div>

          <div class="col-6">
            <q-list bordered separator>
              <q-item
                clickable
                v-ripple
                v-for="(projectItem, projectItemID) in nonNewProjects"
                :key="`project-list-item-${projectItemID}`"
                @click="openGRPCProject(projectID, projectItemID)"
              >
                <q-item-section v-if="projectItem.type === 'grpc'" avatar>
                  <q-icon color="primary" name="folder" />
                </q-item-section>

                <q-item-section v-if="projectItem.type === 'grpc'">gRPC project {{ projectItemID }}</q-item-section>
              </q-item>
            </q-list>
          </div>

          <div class="col-2">
            <q-btn padding="sm" no-caps color="primary" label="New gRPC project" @click="newGRPCProject()" />
          </div>

          <div class="col"></div>
        </div>
      </div>

      <div v-if="project.type === 'grpc'">
        <GRPC :projectID="projectID" />
      </div>
    </q-tab-panel>
  </q-tab-panels>
</template>

<style></style>
