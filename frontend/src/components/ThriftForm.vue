<script>
import { defineComponent } from "vue";
import { VAceEditor } from "vue3-ace-editor";
import "ace-builds/src-noconflict/mode-json";
import "../vendor/merbivore";
import ace from "ace-builds";
import workerJsonUrl from "ace-builds/src-noconflict/worker-json?url";
import { useThriftStore } from "../stores/thrift";

ace.config.setModuleUrl("ace/mode/json_worker", workerJsonUrl);

export default defineComponent({
  name: "ThriftForm",
  components: { VAceEditor },
  props: {
    projectID: String,
    formID: String,
    selectedFunctionID: String,
  },
  data() {
    return { localRequest: "", localHeaders: [] };
  },
  beforeUpdate() {
    this.localRequest = "";
  },
  watch: {
    selectedFunctionID(newValue, oldValue) {
      this.localRequest = "";
    },
  },
  computed: {
    forms() {
      return useThriftStore().projects[this.projectID].forms;
    },
    form() {
      return useThriftStore().projects[this.projectID].forms[this.formID];
    },
    headers() {
      if (this.localHeaders.length > 0) {
        return this.localHeaders;
      }

      return useThriftStore().projects[this.projectID].forms[this.formID].headers;
    },
    address: {
      get() {
        return useThriftStore().projects[this.projectID].forms[this.formID].address;
      },
      async set(address) {
        await useThriftStore().saveAddress(this.projectID, this.formID, address);
      },
    },
    request: {
      get() {
        if (this.localRequest !== "") {
          return this.localRequest;
        }

        return useThriftStore().projects[this.projectID].forms[this.formID].request;
      },
      async set(requestPayload) {
        this.localRequest = requestPayload;
        await useThriftStore().saveRequestPayload(this.projectID, this.formID, requestPayload);
      },
    },
    response: {
      get() {
        let response = useThriftStore().projects[this.projectID].forms[this.formID].response;
        try {
          response = JSON.parse(response);
          response = JSON.stringify(response, null, 4);
        } catch {}

        return response;
      },
      set(value) {
        return (useThriftStore().projects[this.projectID].forms[this.formID].response = value);
      },
    },
  },
  methods: {
    async sendRequest() {
      await useThriftStore().sendRequest(this.projectID, this.formID);
    },

    async stopRequest() {
      await useThriftStore().stopRequest(this.projectID, this.formID);
    },

    async addHeader() {
      await useThriftStore().addHeader(this.projectID, this.formID);
      this.localHeaders = useThriftStore().projects[this.projectID].forms[this.formID].headers;
    },

    async deleteHeader(headerID) {
      await useThriftStore().deleteHeader(this.projectID, this.formID, headerID);
      this.localHeaders = useThriftStore().projects[this.projectID].forms[this.formID].headers;
    },

    async saveHeaders(headers) {
      this.localHeaders = headers;
      await useThriftStore().saveHeaders(this.projectID, this.formID, headers);
    },
  },
});
</script>

<template>
  <div class="full-height">
    <q-form class="q-gutter-md full-height">
      <q-input v-model="address" label="Address" />

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
          :disable="!form.selectedFunctionID"
          @click="sendRequest"
        />
        <q-btn v-else label="Stop" color="negative" @click="stopRequest" />
      </div>
    </q-form>
  </div>
</template>

<style></style>
