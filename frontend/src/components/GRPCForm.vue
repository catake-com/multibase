<script>
import { defineComponent } from "vue";

import { useGRPCStore } from "../stores/grpc";
import { mapState } from "pinia/dist/pinia";

export default defineComponent({
  name: "GRPCForm",
  props: {
    formID: Number,
  },
  computed: {
    ...mapState(useGRPCStore, ["forms"]),
    address: {
      get() {
        return useGRPCStore().forms[this.formID].address;
      },
      set(value) {
        return (useGRPCStore().forms[this.formID].address = value);
      },
    },
    request: {
      get() {
        return useGRPCStore().forms[this.formID].request;
      },
      set(value) {
        return (useGRPCStore().forms[this.formID].request = value);
      },
    },
    response: {
      get() {
        return useGRPCStore().forms[this.formID].response;
      },
      set(value) {
        return (useGRPCStore().forms[this.formID].response = value);
      },
    },
  },
  methods: {
    sendRequest() {
      useGRPCStore().sendRequest(this.formID);
    },

    stopRequest() {
      useGRPCStore().stopRequest(this.formID);
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
