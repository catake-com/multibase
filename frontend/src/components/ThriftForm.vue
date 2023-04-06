<script setup>
import { computed, onBeforeUpdate, ref, watch } from "vue";
import { useThriftStore } from "../stores/thrift";
import { VAceEditor } from "vue3-ace-editor";
import "ace-builds/src-noconflict/mode-json";
import "../vendor/merbivore";
import ace from "ace-builds";
import workerJsonUrl from "ace-builds/src-noconflict/worker-json?url";

ace.config.setModuleUrl("ace/mode/json_worker", workerJsonUrl);

const thriftStore = useThriftStore();

const props = defineProps({
  projectID: String,
  formID: String,
  selectedFunctionID: String,
});

const localRequest = ref("");
const localHeaders = ref([]);

onBeforeUpdate(() => {
  localRequest.value = "";
});

watch(
  () => props.selectedFunctionID,
  async (newValue, oldValue) => {
    localRequest.value = "";
  }
);

const forms = computed(() => thriftStore.project(props.projectID).forms);
const form = computed(() => thriftStore.project(props.projectID).forms[props.formID]);

const headers = computed(() => {
  if (localHeaders.value.length > 0) {
    return localHeaders.value;
  }

  return thriftStore.project(props.projectID).forms[props.formID].headers;
});

const address = computed({
  get() {
    return thriftStore.project(props.projectID).forms[props.formID].address;
  },
  async set(address) {
    await thriftStore.saveAddress(props.projectID, props.formID, address);
  },
});

const isMultiplexed = computed({
  get() {
    return thriftStore.project(props.projectID).forms[props.formID].isMultiplexed;
  },
  async set(isMultiplexed) {
    await thriftStore.saveIsMultiplexed(props.projectID, props.formID, isMultiplexed);
  },
});

const request = computed({
  get() {
    if (localRequest.value !== "") {
      return localRequest.value;
    }

    return thriftStore.project(props.projectID).forms[props.formID].request;
  },
  async set(requestPayload) {
    localRequest.value = requestPayload;
    await thriftStore.saveRequestPayload(props.projectID, props.formID, requestPayload);
  },
});

const response = computed({
  get() {
    let response = thriftStore.project(props.projectID).forms[props.formID].response;
    try {
      response = JSON.parse(response);
      response = JSON.stringify(response, null, 4);
    } catch {}

    return response;
  },
  set(value) {
    return (thriftStore.project(props.projectID).forms[props.formID].response = value);
  },
});

async function sendRequest() {
  await thriftStore.sendRequest(props.projectID, props.formID);
}

async function stopRequest() {
  await thriftStore.stopRequest(props.projectID, props.formID);
}

async function addHeader() {
  await thriftStore.addHeader(props.projectID, props.formID);
  localHeaders.value = thriftStore.project(props.projectID).forms[props.formID].headers;
}

async function beautifyRequest() {
  await thriftStore.beautifyRequest(props.projectID, props.formID);

  localRequest.value = thriftStore.project(props.projectID).forms[props.formID].request;
}

async function deleteHeader(headerID) {
  await thriftStore.deleteHeader(props.projectID, props.formID, headerID);
  localHeaders.value = thriftStore.project(props.projectID).forms[props.formID].headers;
}

async function saveHeaders(headers) {
  localHeaders.value = headers;
  await thriftStore.saveHeaders(props.projectID, props.formID, headers);
}
</script>

<template>
  <div class="full-height">
    <q-form class="q-gutter-md full-height">
      <q-input dense v-model="address" label="Address" debounce="500" />

      <div>
        <q-checkbox v-model="isMultiplexed" label="Enable multiplexed protocol" dense />
      </div>

      <q-btn-group>
        <q-btn outline label="Beautify request JSON" size="xs" @click="beautifyRequest" />

        <q-btn outline label="Add Header" size="xs" @click="addHeader" />
      </q-btn-group>

      <div class="row" v-for="header in headers" :key="header.id">
        <div class="col">
          <q-input dense v-model="header.key" label="Header" @keyup="saveHeaders(headers)" />
        </div>

        <div class="col">
          <q-input dense v-model="header.value" label="Value" @keyup="saveHeaders(headers)" />
        </div>

        <div class="col">
          <q-btn round icon="clear" size="xs" @click="deleteHeader(header.id)" style="margin: 15px 0 0 10px" />
        </div>
      </div>

      <v-ace-editor
        v-model:value="request"
        lang="json"
        theme="merbivore_custom"
        style="height: 30%"
        :options="{ useWorker: true, showPrintMargin: false, behavioursEnabled: false }"
      />

      <v-ace-editor
        v-model:value="response"
        lang="json"
        theme="merbivore_custom"
        style="height: 30%"
        readonly
        :options="{ showPrintMargin: false }"
      />

      <div>
        <q-btn
          v-if="!form.requestInProgress"
          label="Send"
          color="secondary"
          :disable="!form.selectedFunctionID"
          @click="sendRequest"
        />
        <q-btn v-else label="Stop" color="negative" @click="stopRequest" />
      </div>
    </q-form>
  </div>
</template>

<style></style>
