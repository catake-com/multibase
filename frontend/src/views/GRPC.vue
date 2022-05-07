<script>
import { defineComponent } from "vue";
import { mapState } from "pinia";

import { OpenProtoFile, OpenImportPath, SelectMethod, SendRequest } from "../wailsjs/go/main/App";
import { useGRPCStore } from "../stores/grpc";

export default defineComponent({
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
    ...mapState(useGRPCStore, ["importPathList", "nodes"]),
    address: {
      get() {
        return useGRPCStore().forms[useGRPCStore().currentFormID].address;
      },
      set(value) {
        return (useGRPCStore().forms[useGRPCStore().currentFormID].address = value);
      },
    },
    request: {
      get() {
        return useGRPCStore().forms[useGRPCStore().currentFormID].request;
      },
      set(value) {
        return (useGRPCStore().forms[useGRPCStore().currentFormID].request = value);
      },
    },
    response: {
      get() {
        return useGRPCStore().forms[useGRPCStore().currentFormID].response;
      },
      set(value) {
        return (useGRPCStore().forms[useGRPCStore().currentFormID].response = value);
      },
    },
  },
  watch: {
    selectedMethod(newMethod, oldMethod) {
      if (newMethod === oldMethod) {
        return;
      }

      const currentMethod = newMethod || oldMethod;

      SelectMethod(currentMethod)
        .then((payload) => {
          this.request = payload;
        })
        .catch((reason) => {
          this.response = reason;
        });
    },
  },
  methods: {
    sendRequest() {
      SendRequest(this.address, this.selectedMethod, this.request)
        .then((response) => {
          this.response = response;
        })
        .catch((reason) => {
          this.response = reason;
        });

      useGRPCStore().saveState();
    },

    openProtoFile() {
      const store = useGRPCStore();

      OpenProtoFile()
        .then((result) => {
          if (result.protoFilePath !== "") {
            store.addProtoFile(result.protoFilePath, result.currentDir);
          }
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
        <q-form @submit="sendRequest" class="q-gutter-md">
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
