<script>
import { defineComponent } from "vue";

import { useGRPCStore } from "../stores/grpc";

export default defineComponent({
  name: "GRPCForm",
  props: {
    formID: Number,
  },
  computed: {
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
  },
});
</script>

<template>
  <div>
    <q-form @submit="sendRequest" class="q-gutter-md">
      <q-input v-model="address" label="Address" />

      <q-input type="textarea" v-model="request" label="Request" />

      <q-input type="textarea" v-model="response" label="Response" />

      <div>
        <q-btn label="Send" type="submit" color="primary" />
      </div>
    </q-form>
  </div>
</template>

<style></style>
