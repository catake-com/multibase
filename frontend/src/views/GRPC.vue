<script>
import { defineComponent } from "vue";
import { mapState, mapWritableState } from "pinia";

import { OpenProtoFile, OpenImportPath } from "../wailsjs/go/main/App";
import { useGRPCStore } from "../stores/grpc";

export default defineComponent({
  data() {
    return {
      val: 30,
      selected: null,
      tab: "protos",
    };
  },
  beforeRouteEnter() {
    useGRPCStore().loadState();
  },
  computed: {
    ...mapState(useGRPCStore, ["importPathList", "nodes"]),
    ...mapWritableState(useGRPCStore, ["address", "request", "response"]),
  },
  methods: {
    onSubmit() {
      console.log("submit");
    },

    openProtoFile() {
      const store = useGRPCStore();

      OpenProtoFile()
        .then((path) => {
          store.addProtoFile(path);
        })
        .catch((reason) => {
          this.response = reason;
        });
    },

    openImportPath() {
      const store = useGRPCStore();

      OpenImportPath()
        .then((path) => {
          store.addImportPath(path);
        })
        .catch((reason) => {
          this.response = reason;
        });
    },

    removeImportPath(importPath) {
      const store = useGRPCStore();

      store.removeImportPath(importPath);
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

            <q-tree :nodes="nodes" default-expand-all v-model:selected="selected" node-key="label" />
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
        <q-form @submit="onSubmit" class="q-gutter-md">
          <q-input v-model="address" label="Address" />

          <q-input type="textarea" v-model="request" label="Request" />

          <q-input type="textarea" v-model="response" label="Response" />

          <div>
            <q-btn label="Send" type="submit" color="primary" />
          </div>
        </q-form>
      </template>
    </q-splitter>
  </div>
</template>

<style></style>
