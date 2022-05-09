<script>
import { defineComponent } from "vue";
import { mapState, mapWritableState } from "pinia";

import { OpenProtoFile, OpenImportPath } from "../wailsjs/go/main/App";
import { useGRPCStore } from "../stores/grpc";
import GRPCForm from "../components/GRPCForm.vue";

export default defineComponent({
  name: "GRPC",
  components: { GRPCForm },
  data() {
    return {
      val: 30,
      selectedMethod: null,
      tab: "protos",
    };
  },
  beforeRouteEnter() {
    useGRPCStore().loadState();
  },
  computed: {
    ...mapState(useGRPCStore, ["importPathList", "nodes", "forms"]),
    ...mapWritableState(useGRPCStore, ["currentFormID"]),
  },
  watch: {
    selectedMethod(newMethod, oldMethod) {
      if (newMethod === oldMethod) {
        return;
      }

      const currentMethod = newMethod || oldMethod;

      useGRPCStore().selectMethod(this.currentFormID, currentMethod);
    },
  },
  methods: {
    openProtoFile() {
      const store = useGRPCStore();

      OpenProtoFile()
        .then((result) => {
          if (result.protoFilePath !== "") {
            store.addProtoFile(result.protoFilePath, result.currentDir);
          }
        })
        .catch((reason) => {
          store.forms[this.currentFormID].response = reason;
        });
    },

    openImportPath() {
      const store = useGRPCStore();

      OpenImportPath()
        .then((path) => {
          store.addImportPath(path);
        })
        .catch((reason) => {
          store.forms[this.currentFormID].response = reason;
        });
    },

    removeImportPath(importPath) {
      const store = useGRPCStore();

      store.removeImportPath(importPath);
    },

    createNewForm() {
      const store = useGRPCStore();

      store.createNewForm();
    },
  },
});
</script>

<template>
  <div>
    <q-splitter v-model="val">
      <template v-slot:before>
        <q-tabs v-model="tab">
          <q-tab name="protos" label="Protos" />
          <q-tab name="import_paths" label="Import Paths" />
        </q-tabs>

        <q-separator />

        <q-tab-panels v-model="tab" animated>
          <q-tab-panel name="protos">
            <q-btn size="sm" label="Open .proto file" @click="openProtoFile" />

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
            <q-btn size="sm" label="Add import path" @click="openImportPath" />

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
          <q-tab
            :name="parseInt(formID)"
            :label="form.selectedMethodID || 'New Form'"
            v-for="(form, formID) in forms"
            :key="`tab-${formID}`"
          />

          <q-btn @click="createNewForm" label="+" color="secondary" />
        </q-tabs>

        <q-separator />

        <q-tab-panels v-model="currentFormID" animated>
          <q-tab-panel :name="parseInt(formID)" v-for="(form, formID) in forms" :key="`tab-panel-${formID}`">
            <GRPCForm :formID="parseInt(formID)" />
          </q-tab-panel>
        </q-tab-panels>
      </template>
    </q-splitter>
  </div>
</template>

<style></style>
