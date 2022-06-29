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
    nonNewProjectList() {
      return Object.values(this.nonNewProjects);
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

    async newGRPCProject() {
      await useGRPCStore().createNewProject(this.currentProjectID);
      await useProjectStore().createNewGRPCProject(this.currentProjectID);
    },

    deleteGRPCProject() {
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
            <q-table
              :rows="nonNewProjectList"
              :columns="[
                { name: 'icon', field: 'icon', align: 'left' },
                { name: 'name', field: 'name', align: 'left' },
                { name: 'actions', field: 'actions', align: 'right' },
              ]"
              hide-header
              hide-pagination
              dense
              bordered
              separator="cell"
              no-data-label="No projects has been created yet"
              row-key="id"
            >
              <template v-slot:body="props">
                <q-tr :props="props">
                  <q-td key="icon" :props="props" auto-width no-hover>
                    <q-icon v-if="props.row.type === 'grpc'" name="img:grpc.jpg" size="36px" />
                  </q-td>

                  <q-td
                    key="name"
                    :props="props"
                    style="cursor: pointer"
                    @click="openGRPCProject(projectID, props.row.id)"
                  >
                    {{ props.row.name }}
                  </q-td>

                  <q-td key="actions" :props="props" auto-width no-hover>
                    <q-btn
                      class="inline"
                      icon="delete"
                      size="10px"
                      style="width: 20px"
                      flat
                      rounded
                      dense
                      @click="deleteGRPCProject($event, props.row.id)"
                    />
                  </q-td>
                </q-tr>
              </template>

              <template v-slot:no-data="{ icon, message, filter }">
                <div class="full-width row flex-center">
                  {{ message }}
                </div>
              </template>
            </q-table>
          </div>

          <div class="col-2">
            <q-btn padding="sm" no-caps color="primary" label="New gRPC project" @click="newGRPCProject()" />
          </div>

          <div class="col"></div>
        </div>
      </div>

      <div v-if="project.type === 'grpc'" class="full-height">
        <GRPC :projectID="projectID" />
      </div>
    </q-tab-panel>
  </q-tab-panels>
</template>

<style></style>
