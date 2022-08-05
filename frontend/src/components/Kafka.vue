<script>
import { defineComponent } from "vue";
import { mapState } from "pinia";

import { useKafkaStore } from "../stores/kafka";

export default defineComponent({
  name: "Kafka",
  props: {
    projectID: String,
  },
  beforeCreate() {
    useKafkaStore().loadState();
  },
  computed: {
    ...mapState(useKafkaStore, ["projects"]),
  },
  methods: {},
});
</script>

<template>
  <div class="full-height">kafka</div>
</template>

<style></style>
