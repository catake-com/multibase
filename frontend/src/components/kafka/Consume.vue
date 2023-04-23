<script setup>
import { computed } from "vue";
import { useQuasar } from "quasar";
import { useKafkaStore } from "../../stores/kafka";

const quasar = useQuasar();

const kafkaStore = useKafkaStore();

const props = defineProps({
  projectID: String,
});

const splitterWidthConsuming = 12;

const hoursAgo = computed({
  get() {
    return kafkaStore.consumingSession(props.projectID).hoursAgo;
  },
  async set(hoursAgo) {
    kafkaStore.consumingSession(props.projectID).hoursAgo = hoursAgo;
  },
});

const currentTopic = computed(() => kafkaStore.consumingSession(props.projectID).currentTopic);
const consumedTopic = computed(() => kafkaStore.consumedTopic(props.projectID));
const consumedTopicMessages = computed(() => kafkaStore.consumedTopicMessages(props.projectID));

async function stopTopicConsuming() {
  try {
    await kafkaStore.stopTopicConsuming(props.projectID);
    quasar.notify({ type: "positive", message: "Stopped consuming" });
  } catch (error) {
    quasar.notify({ type: "negative", message: error });
  }
}

async function restartTopicConsuming() {
  try {
    await kafkaStore.restartTopicConsuming(props.projectID);
    quasar.notify({ type: "positive", message: "Restarted consuming" });
  } catch (error) {
    quasar.notify({ type: "negative", message: error });
  }
}

function consumedMessagesTableRowKey(row) {
  return `${row.partitionID}_${row.offset}`;
}

const consumedMessagesTableColumns = [
  {
    name: "timestamp",
    label: "Timestamp",
    align: "left",
    field: "timestamp",
  },
  { name: "partition", align: "left", label: "Partition", field: "partitionID" },
  { name: "offset", align: "left", label: "Offset", field: "offset" },
  { name: "key", align: "left", label: "Key", field: "key" },
  { name: "data", align: "left", label: "Data", field: "data" },
  { name: "headers", align: "left", label: "Headers", field: "" },
];

const consumedMessagesTablePagination = {
  rowsPerPage: 20,
};

const consumedMessagesTableRowsPerPage = [5, 10, 20, 50, 100, 200, 500];
</script>

<template>
  <q-btn label="Stop" color="secondary" @click="stopTopicConsuming()" />

  <q-splitter v-model="splitterWidthConsuming" class="full-height" disable>
    <template v-slot:before>
      {{ currentTopic }}

      <q-input v-model="hoursAgo" label="Show messages for last X hours" />

      <!--          <div>Count total: {{ consumedTopic?.countTotal }}</div>-->
      <!--          Partitions:-->
      <!--          <q-list>-->
      <!--            <q-item v-for="partition in consumedTopic?.partitions" :key="partition.id">-->
      <!--              <q-item-section>-->
      <!--                <q-item-label overline>{{ partition.id }}</q-item-label>-->
      <!--                <q-item-label>{{ partition.offsetTotalStart }} - {{ partition.offsetTotalEnd }}</q-item-label>-->
      <!--              </q-item-section>-->
      <!--            </q-item>-->
      <!--          </q-list>-->

      <q-btn label="Refresh" color="secondary" @click="restartTopicConsuming()" />
    </template>

    <template v-slot:after>
      <q-table
        :rows="consumedTopicMessages"
        :columns="consumedMessagesTableColumns"
        :row-key="consumedMessagesTableRowKey"
        :pagination="consumedMessagesTablePagination"
        :rows-per-page-options="consumedMessagesTableRowsPerPage"
      />
    </template>
  </q-splitter>
</template>

<style></style>
