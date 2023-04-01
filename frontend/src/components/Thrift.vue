<script setup>
import { computed, watch } from "vue";
import { useThriftStore } from "../stores/thrift";
import ThriftForm from "./ThriftForm.vue";

const thriftStore = useThriftStore();

const props = defineProps({
  projectID: String,
});

await thriftStore.loadProject(props.projectID);

const nodes = computed(() => thriftStore.project(props.projectID).nodes);
const forms = computed(() => thriftStore.project(props.projectID).forms);
const formIDs = computed(() => thriftStore.project(props.projectID).formIDs);

const currentFormID = computed({
  get() {
    return thriftStore.project(props.projectID).currentFormID;
  },
  async set(currentFormID) {
    await thriftStore.saveCurrentFormID(props.projectID, currentFormID);
  },
});

const splitterWidth = computed({
  get() {
    return thriftStore.project(props.projectID).splitterWidth;
  },
  async set(splitterWidth) {
    await thriftStore.saveSplitterWidth(props.projectID, splitterWidth);
  },
});

const selectedFunction = computed({
  get() {
    const currentForm = thriftStore.project(props.projectID).forms[thriftStore.project(props.projectID).currentFormID];

    if (currentForm) {
      return currentForm.selectedFunctionID;
    }
  },
  async set(selectedFunctionID) {
    await thriftStore.selectFunction(
      props.projectID,
      thriftStore.project(props.projectID).currentFormID,
      selectedFunctionID
    );
  },
});

watch(
  () => thriftStore.project(props.projectID).currentFormID,
  async (newCurrentFormID, oldCurrentFormID) => {
    if (newCurrentFormID === oldCurrentFormID) {
      return;
    }

    const formID = newCurrentFormID || oldCurrentFormID;
    const form = thriftStore.project(props.projectID).forms[formID];

    if (form.selectedFunctionID && selectedFunction.value !== form.selectedFunctionID) {
      selectedFunction.value = form.selectedFunctionID;
    }
  }
);

async function openFilePath() {
  await thriftStore.openFilePath(props.projectID);
  console.log(thriftStore.project(props.projectID));
  console.log(nodes);
}

async function createNewForm() {
  await thriftStore.createNewForm(props.projectID);
}

async function closeFormTab(event, formID) {
  event.preventDefault();
  await thriftStore.removeForm(props.projectID, formID);
}
</script>

<template>
  <div class="full-height">
    <q-splitter v-if="splitterWidth" v-model="splitterWidth" class="full-height" :limits="[20, 80]">
      <template v-slot:before>
        <q-btn size="sm" label="Open Thrift file" color="primary" @click="openFilePath" />

        <q-tree
          v-if="(nodes || []).length > 0"
          ref="serviceTree"
          :nodes="nodes"
          default-expand-all
          no-selection-unset
          v-model:selected="selectedFunction"
          node-key="id"
        />
      </template>

      <template v-slot:after>
        <q-tabs v-model="currentFormID" align="left" outside-arrows mobile-arrows dense no-caps>
          <q-tab :name="formID" v-for="formID in formIDs" :key="`tab-${formID}`">
            <div class="row justify-between">
              <div class="col q-tab__label">
                <div v-if="forms[formID].selectedFunctionID.length < 15">
                  {{ forms[formID].selectedFunctionID || "New Form" }}
                </div>

                <div v-else class="thrift-form-tab-name">
                  <div class="start">{{ forms[formID].selectedFunctionID.substring(0, 20) }}</div>
                  <div class="end">
                    {{
                      forms[formID].selectedFunctionID.substring(
                        forms[formID].selectedFunctionID.length - 20 > 20
                          ? forms[formID].selectedFunctionID.length - 20
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
                  :disable="Object.keys(this.forms).length === 1"
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
            <ThriftForm :formID="formID" :projectID="this.projectID" :selectedFunctionID="this.selectedFunction" />
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

.thrift-form-tab-name {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  justify-content: flex-start;
}

.thrift-form-tab-name > .start {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex-shrink: 1;
}

.thrift-form-tab-name > .end {
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
