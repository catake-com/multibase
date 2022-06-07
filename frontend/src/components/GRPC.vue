<script>
import { defineComponent } from "vue";
import { mapState } from "pinia";

import { useGRPCStore } from "../stores/grpc";
import GRPCForm from "./GRPCForm.vue";

export default defineComponent({
  name: "GRPC",
  props: {
    projectID: String,
  },
  components: { GRPCForm },
  data() {
    return {
      val: 30,
      selectedMethod: null,
      tab: "protos",
    };
  },
  beforeCreate() {
    // TODO: load state per project instead of all state
    useGRPCStore().loadState();
  },
  computed: {
    ...mapState(useGRPCStore, ["projects"]),
    importPathList() {
      if (useGRPCStore().projects[this.projectID]) {
        return useGRPCStore().projects[this.projectID].importPathList;
      }
    },
    nodes() {
      if (useGRPCStore().projects[this.projectID]) {
        return useGRPCStore().projects[this.projectID].nodes;
      }
    },
    forms() {
      if (useGRPCStore().projects[this.projectID]) {
        return useGRPCStore().projects[this.projectID].forms;
      }
    },
    currentFormID: {
      get() {
        if (useGRPCStore().projects[this.projectID]) {
          return useGRPCStore().projects[this.projectID].currentFormID;
        }
      },
      set(value) {
        return (useGRPCStore().projects[this.projectID].currentFormID = value);
      },
    },
  },
  watch: {
    selectedMethod(newMethod, oldMethod) {
      if (newMethod === oldMethod) {
        return;
      }

      const currentMethod = newMethod || oldMethod;

      useGRPCStore().selectMethod(this.projectID, this.projects[this.projectID].currentFormID, currentMethod);
    },
  },
  methods: {
    openProtoFile() {
      const store = useGRPCStore();

      store.openProtoFile(this.projectID);
    },

    openImportPath() {
      const store = useGRPCStore();

      store.openImportPath(this.projectID);
    },

    removeImportPath(importPath) {
      const store = useGRPCStore();

      store.removeImportPath(this.projectID, importPath);
    },

    createNewForm() {
      const store = useGRPCStore();

      store.createNewForm(this.projectID);
    },

    closeFormTab(event, formID) {
      event.preventDefault();

      const store = useGRPCStore();

      store.removeForm(this.projectID, formID);
    },
  },
});
</script>

<template>
  <div class="full-height">
    <q-splitter v-model="val" class="full-height">
      <template v-slot:before>
        <q-tabs v-model="tab">
          <q-tab name="protos" label="Protos" />
          <q-tab name="import_paths" label="Import Paths" />
        </q-tabs>

        <q-separator />

        <q-tab-panels v-model="tab" animated>
          <q-tab-panel name="protos">
            <q-btn size="sm" label="Open .proto file" color="primary" @click="openProtoFile" />

            <q-tree
              v-if="(nodes || []).length > 0"
              :nodes="nodes"
              default-expand-all
              no-selection-unset
              v-model:selected="selectedMethod"
              node-key="id"
            />
          </q-tab-panel>

          <q-tab-panel name="import_paths">
            <q-btn size="sm" label="Add import path" color="primary" @click="openImportPath" />

            <q-list dense>
              <q-item v-for="importPath in importPathList" :key="importPath">
                <q-item-section avatar>
                  <q-icon name="folder" />
                </q-item-section>

                <q-item-section>
                  <span>{{ importPath }}</span>
                </q-item-section>

                <q-item-section avatar>
                  <q-icon name="delete" @click="removeImportPath(importPath)" />
                </q-item-section>
              </q-item>
            </q-list>
          </q-tab-panel>
        </q-tab-panels>
      </template>

      <template v-slot:after>
        <q-tabs v-model="currentFormID" align="left" outside-arrows mobile-arrows dense no-caps>
          <q-tab :name="formID" v-for="(form, formID) in forms" :key="`tab-${formID}`">
            <div class="row justify-between">
              <div class="col q-tab__label">
                <div v-if="form.selectedMethodID.length < 15">{{ form.selectedMethodID || "New Form" }}</div>

                <div v-else class="grpc-form-tab-name">
                  <div class="start">{{ form.selectedMethodID.substring(0, 20) }}</div>
                  <div class="end">
                    {{
                      form.selectedMethodID.substring(
                        form.selectedMethodID.length - 20 > 20 ? form.selectedMethodID.length - 20 : 20
                      )
                    }}
                  </div>
                </div>
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
                  :disable="Object.keys(this.forms).length === 1"
                  @click="closeFormTab($event, formID)"
                />
              </div>
            </div>
          </q-tab>

          <q-btn @click="createNewForm" icon="add" color="primary" />
        </q-tabs>

        <q-separator />

        <q-tab-panels v-model="currentFormID" animated class="full-height">
          <q-tab-panel :name="formID" v-for="(form, formID) in forms" :key="`tab-panel-${formID}`">
            <GRPCForm :formID="formID" :projectID="this.projectID" />
          </q-tab-panel>
        </q-tab-panels>
      </template>
    </q-splitter>
  </div>
</template>

<style>
.grpc-form-tab-name {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  justify-content: flex-start;
}

.grpc-form-tab-name > .start {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex-shrink: 1;
}

.grpc-form-tab-name > .end {
  white-space: nowrap;
  flex-basis: content;
  flex-grow: 0;
  flex-shrink: 0;
}
</style>
