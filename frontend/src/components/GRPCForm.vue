<script>
import { defineComponent } from "vue";

import { useGRPCStore } from "../stores/grpc";

export default defineComponent({
  name: "GRPCForm",
  props: {
    projectID: String,
    formID: String,
  },
  computed: {
    forms() {
      return useGRPCStore().projects[this.projectID].forms;
    },
    address: {
      get() {
        return useGRPCStore().projects[this.projectID].forms[this.formID].address;
      },
      set(value) {
        return (useGRPCStore().projects[this.projectID].forms[this.formID].address = value);
      },
    },
    request: {
      get() {
        return useGRPCStore().projects[this.projectID].forms[this.formID].request;
      },
      set(value) {
        return (useGRPCStore().projects[this.projectID].forms[this.formID].request = value);
      },
    },
    response: {
      get() {
        return useGRPCStore().projects[this.projectID].forms[this.formID].response;
      },
      set(value) {
        return (useGRPCStore().projects[this.projectID].forms[this.formID].response = value);
      },
    },
  },
  methods: {
    sendRequest() {
      useGRPCStore().sendRequest(this.projectID, this.formID);
    },

    stopRequest() {
      useGRPCStore().stopRequest(this.projectID, this.formID);
    },
  },
});
</script>

<template>
  <div class="full-height">
    <q-form class="q-gutter-md full-height">
      <q-input v-model="address" label="Address" />

      <q-input type="textarea" v-model="request" label="Request" filled />

      <q-input type="textarea" v-model="response" label="Response" filled />

      <div>
        <q-btn v-if="!this.forms[this.formID].requestInProgress" label="Send" color="secondary" @click="sendRequest" />
        <q-btn v-else label="Stop" color="negative" @click="stopRequest" />
      </div>
    </q-form>
  </div>
</template>

<style></style>
