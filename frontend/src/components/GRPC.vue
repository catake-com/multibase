<script setup>
import { computed, ref, watch } from "vue";
import { useGRPCStore } from "../stores/grpc";
import GRPCForm from "./GRPCForm.vue";

const grpcStore = useGRPCStore();

const props = defineProps({
  projectID: String,
});

const tab = ref("protos");

await grpcStore.loadProject(props.projectID);

const importPathList = computed(() => grpcStore.project(props.projectID).importPathList);
const nodes = computed(() => grpcStore.project(props.projectID).nodes);
const forms = computed(() => grpcStore.project(props.projectID).forms);
const formIDs = computed(() => grpcStore.project(props.projectID).formIDs);

const currentFormID = computed({
  get() {
    return grpcStore.project(props.projectID).currentFormID;
  },
  async set(currentFormID) {
    await grpcStore.saveCurrentFormID(props.projectID, currentFormID);
  },
});

const splitterWidth = computed({
  get() {
    return grpcStore.project(props.projectID).splitterWidth;
  },
  async set(splitterWidth) {
    await grpcStore.saveSplitterWidth(props.projectID, splitterWidth);
  },
});

const selectedMethod = computed({
  get() {
    const currentForm = grpcStore.project(props.projectID).forms[grpcStore.project(props.projectID).currentFormID];

    if (currentForm) {
      return currentForm.selectedMethodID;
    }
  },
  async set(selectedMethodID) {
    await grpcStore.selectMethod(props.projectID, grpcStore.project(props.projectID).currentFormID, selectedMethodID);
  },
});

watch(
  () => grpcStore.project(props.projectID).currentFormID,
  async (newCurrentFormID, oldCurrentFormID) => {
    if (newCurrentFormID === oldCurrentFormID) {
      return;
    }

    const formID = newCurrentFormID || oldCurrentFormID;
    const form = grpcStore.project(props.projectID).forms ? grpcStore.project(props.projectID).forms[formID] : null;

    if (form && form.selectedMethodID && selectedMethod.value !== form.selectedMethodID) {
      selectedMethod.value = form.selectedMethodID;
    }
  }
);

async function openProtoFile() {
  await grpcStore.openProtoFile(props.projectID);
}

async function deleteAllProtoFiles() {
  await grpcStore.deleteAllProtoFiles(props.projectID);
}

async function openImportPath() {
  await grpcStore.openImportPath(props.projectID);
}

async function removeImportPath(importPath) {
  await grpcStore.removeImportPath(props.projectID, importPath);
}

async function createNewForm() {
  await grpcStore.createNewForm(props.projectID);
}

async function closeFormTab(event, formID) {
  event.preventDefault();
  await grpcStore.removeForm(props.projectID, formID);
}
</script>

<template>
  <div class="full-height">
    <q-splitter v-if="splitterWidth" v-model="splitterWidth" class="full-height" :limits="[20, 80]">
      <template v-slot:before>
        <q-tabs dense v-model="tab" class="full-width">
          <q-tab name="protos" label="Protos" />
          <q-tab name="import_paths" label="Import Paths" />
        </q-tabs>

        <q-separator />

        <q-tab-panels v-model="tab" animated style="height: calc(100% - 49px) !important">
          <q-tab-panel name="protos" class="full-height">
            <q-btn size="sm" label="Open .proto file" color="primary" @click="openProtoFile" />
            <q-btn size="sm" icon="delete" color="primary" round @click="deleteAllProtoFiles" class="float-right" />

            <q-tree
              v-if="(nodes || []).length > 0"
              ref="protoTree"
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
          <q-tab :name="formID" v-for="formID in formIDs" :key="`tab-${formID}`">
            <div class="row justify-between">
              <div class="col q-tab__label">
                <div v-if="forms[formID].selectedMethodID.length < 15">
                  {{ forms[formID].selectedMethodID || "New Form" }}
                </div>

                <div v-else class="grpc-form-tab-name">
                  <div class="start">{{ forms[formID].selectedMethodID.substring(0, 20) }}</div>
                  <div class="end">
                    {{
                      forms[formID].selectedMethodID.substring(
                        forms[formID].selectedMethodID.length - 20 > 20
                          ? forms[formID].selectedMethodID.length - 20
                          : 20
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
                  :disable="forms && Object.keys(forms).length === 1"
                  @click="closeFormTab($event, formID)"
                />
              </div>
            </div>
          </q-tab>

          <q-btn @click="createNewForm" icon="add" color="primary" />
        </q-tabs>

        <q-separator />

        <q-tab-panels id="formContainer" v-model="currentFormID" animated>
          <q-tab-panel :name="formID" v-for="(form, formID) in forms" :key="`tab-panel-${formID}`">
            <GRPCForm :formID="formID" :projectID="props.projectID" :selectedMethodID="selectedMethod" />
          </q-tab-panel>
        </q-tab-panels>
      </template>
    </q-splitter>
  </div>
</template>

<style>
#formContainer {
  height: calc(100% - 48px) !important;
}

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

.q-tree__node--selected .q-tree__node-header-content {
  color: #3498db;
}

.q-tabs__content--align-center .q-tab {
  flex: 1 1 auto;
}
</style>
