<script>
import { defineComponent } from "vue";

import { useGRPCStore } from "../stores/grpc";

export default defineComponent({
  name: "GRPCForm",
  props: {
    projectID: Number,
    formID: Number,
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
  <div>
    <q-form class="q-gutter-md">
      <q-input v-model="address" label="Address" />

      <q-input type="textarea" v-model="request" label="Request" />

      <q-input type="textarea" v-model="response" label="Response" />

      <div>
        <q-btn v-if="!this.forms[this.formID].requestInProgress" label="Send" color="primary" @click="sendRequest" />
        <q-btn v-else label="Stop" color="negative" @click="stopRequest" />
      </div>
    </q-form>
  </div>
</template>

<style></style>
