<script setup>
import { computed, onBeforeUpdate, ref, watch } from "vue";
import { useGRPCStore } from "../stores/grpc";
import { VAceEditor } from "vue3-ace-editor";
import "ace-builds/src-noconflict/mode-json";
import "../vendor/merbivore";
import ace from "ace-builds";
import workerJsonUrl from "ace-builds/src-noconflict/worker-json?url";

ace.config.setModuleUrl("ace/mode/json_worker", workerJsonUrl);

const grpcStore = useGRPCStore();

const props = defineProps({
  projectID: String,
  formID: String,
  selectedMethodID: String,
});

const localRequest = ref("");
const localHeaders = ref([]);

onBeforeUpdate(() => {
  localRequest.value = "";
});

watch(
  () => props.selectedMethodID,
  async (newValue, oldValue) => {
    localRequest.value = "";
  }
);

const forms = grpcStore.project(props.projectID).forms;
const form = grpcStore.project(props.projectID).forms[props.formID];

const headers = computed(() => {
  if (localHeaders.value.length > 0) {
    return localHeaders.value;
  }

  return grpcStore.project(props.projectID).forms[props.formID].headers;
});

const address = computed({
  get() {
    return grpcStore.project(props.projectID).forms[props.formID].address;
  },
  async set(address) {
    await grpcStore.saveAddress(props.projectID, props.formID, address);
  },
});

const request = computed({
  get() {
    if (localRequest.value !== "") {
      return localRequest.value;
    }

    return grpcStore.project(props.projectID).forms[props.formID].request;
  },
  async set(requestPayload) {
    localRequest.value = requestPayload;
    await grpcStore.saveRequestPayload(props.projectID, props.formID, requestPayload);
  },
});

const response = computed({
  get() {
    let response = grpcStore.project(props.projectID).forms[props.formID].response;
    try {
      response = JSON.parse(response);
      response = JSON.stringify(response, null, 4);
    } catch {}

    return response;
  },
  set(value) {
    return (grpcStore.project(props.projectID).forms[props.formID].response = value);
  },
});

async function sendRequest() {
  await grpcStore.sendRequest(props.projectID, props.formID);
}

async function stopRequest() {
  await grpcStore.stopRequest(props.projectID, props.formID);
}

async function reflectProto() {
  await grpcStore.reflectProto(props.projectID, props.formID);
}

async function addHeader() {
  await grpcStore.addHeader(props.projectID, props.formID);
  localHeaders.value = grpcStore.project(props.projectID).forms[props.formID].headers;
}

async function deleteHeader(headerID) {
  await grpcStore.deleteHeader(props.projectID, props.formID, headerID);
  localHeaders.value = grpcStore.project(props.projectID).forms[props.formID].headers;
}

async function saveHeaders(headers) {
  localHeaders.value = headers;
  await grpcStore.saveHeaders(props.projectID, props.formID, headers);
}
</script>

<template>
  <div class="full-height">
    <q-form class="q-gutter-md full-height">
      <q-input v-model="address" label="Address" debounce="500" />

      <div>
        <q-btn outline label="Import proto from server reflection" size="xs" @click="reflectProto" />
      </div>

      <q-btn outline label="Add Header" size="xs" @click="addHeader" />

      <div class="row" v-for="header in headers" :key="header.id">
        <div class="col">
          <q-input v-model="header.key" label="Header" @keyup="saveHeaders(headers)" />
        </div>

        <div class="col">
          <q-input v-model="header.value" label="Value" @keyup="saveHeaders(headers)" />
        </div>

        <div class="col">
          <q-btn round icon="delete" size="xs" @click="deleteHeader(header.id)" style="margin: 30px 0 0 10px" />
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
          :disable="!form.selectedMethodID"
          @click="sendRequest"
        />
        <q-btn v-else label="Stop" color="negative" @click="stopRequest" />
      </div>
    </q-form>
  </div>
</template>

<style></style>
