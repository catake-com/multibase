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
  },
  data() {
    return { localRequest: "" };
  },
  beforeUpdate() {
    this.localRequest = "";
  },
  computed: {
    forms() {
      return useThriftStore().projects[this.projectID].forms;
    },
    form() {
      return useThriftStore().projects[this.projectID].forms[this.formID];
    },
    address: {
      get() {
        return useThriftStore().projects[this.projectID].forms[this.formID].address;
      },
      async set(address) {
        await useThriftStore().saveAddress(this.projectID, this.formID, address);
      },
    },g
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
  },
});
</script>

<template>
  <div class="full-height">
    <q-form class="q-gutter-md full-height">
      <q-input v-model="address" label="Address" />

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
