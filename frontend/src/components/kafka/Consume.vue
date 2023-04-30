<script setup>
import { computed, ref } from "vue";
import { date, useQuasar } from "quasar";
import { useKafkaStore } from "../../stores/kafka";

const quasar = useQuasar();

const kafkaStore = useKafkaStore();

const props = defineProps({
  projectID: String,
});

const splitterWidthConsuming = 20;

const currentConsumedTopic = computed(() => kafkaStore.initiatedTopicConsuming(props.projectID).topicName);
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
    await kafkaStore.restartTopicConsuming(props.projectID, startFromTime.value);
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

const currentTimeTenMinutesAgo = date.subtractFromDate(Date.now(), { minutes: 10 });

const startFrom = ref("date");
const startFromTime = ref(date.formatDate(currentTimeTenMinutesAgo, "YYYY-MM-DD HH:mm:ss Z"));

function setCurrentTimeMinutes(minutes) {
  const currentTimeMinutesAgo = date.subtractFromDate(Date.now(), { minutes: minutes });
  startFromTime.value = date.formatDate(currentTimeMinutesAgo, "YYYY-MM-DD HH:mm:ss Z");
}

try {
  kafkaStore.startTopicConsuming(props.projectID, currentConsumedTopic.value, startFromTime.value);
} catch (error) {
  quasar.notify({ type: "negative", message: error });
}
</script>

<template>
  <q-btn label="Stop" color="secondary" @click="stopTopicConsuming()" />

  <q-splitter v-model="splitterWidthConsuming" class="full-height" disable>
    <template v-slot:before>
      <div class="text-subtitle1">Start from:</div>

      <q-btn-toggle
        v-model="startFrom"
        size="sm"
        toggle-color="primary"
        :options="[
          { label: 'Date', value: 'date' },
          { label: 'Offset', value: 'offset' },
        ]"
      />

      <div v-if="startFrom === 'date'">
        <div class="row no-wrap">
          <q-btn-dropdown
            outline
            color="info"
            :label="startFromTime"
            dropdown-icon="none"
            @hide="restartTopicConsuming()"
          >
            <div class="row items-start">
              <q-date v-model="startFromTime" mask="YYYY-MM-DD HH:mm:ss Z" color="primary" first-day-of-week="1" />
              <q-time v-model="startFromTime" mask="YYYY-MM-DD HH:mm:ss Z" color="primary" with-seconds format24h />
            </div>
          </q-btn-dropdown>

          <q-btn-dropdown outline color="info" @hide="restartTopicConsuming()">
            <q-list>
              <q-item clickable v-close-popup @click="setCurrentTimeMinutes(10)">
                <q-item-section>
                  <q-item-label>Last 10 minutes</q-item-label>
                </q-item-section>
              </q-item>

              <q-item clickable v-close-popup @click="setCurrentTimeMinutes(30)">
                <q-item-section>
                  <q-item-label>Last 30 minutes</q-item-label>
                </q-item-section>
              </q-item>

              <q-item clickable v-close-popup @click="setCurrentTimeMinutes(60)">
                <q-item-section>
                  <q-item-label>Last 1 hour</q-item-label>
                </q-item-section>
              </q-item>

              <q-item clickable v-close-popup @click="setCurrentTimeMinutes(60 * 24)">
                <q-item-section>
                  <q-item-label>Last 24 hours</q-item-label>
                </q-item-section>
              </q-item>
            </q-list>
          </q-btn-dropdown>
        </div>
      </div>

      <div v-if="startFrom === 'offset'">offset</div>

      {{ currentConsumedTopic }}

      <div>Count total: {{ consumedTopic?.countTotal }}</div>
      Partitions:
      <q-list>
        <q-item v-for="partition in consumedTopic?.partitions" :key="partition.id">
          <q-item-section>
            <q-item-label overline>{{ partition.id }}</q-item-label>
            <q-item-label>{{ partition.offsetTotalStart }} - {{ partition.offsetTotalEnd }}</q-item-label>
          </q-item-section>
        </q-item>
      </q-list>

      <q-btn label="Refresh" color="secondary" @click="restartTopicConsuming()" />
    </template>

    <template v-slot:after>
      <q-table
        :rows="consumedTopicMessages"
        :columns="consumedMessagesTableColumns"
        :row-key="consumedMessagesTableRowKey"
        :pagination="consumedMessagesTablePagination"
        :rows-per-page-options="consumedMessagesTableRowsPerPage"
        :loading="!consumedTopic.topicName"
      />
    </template>
  </q-splitter>
</template>

<style></style>
