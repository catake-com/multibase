<script setup>
import { computed, ref, watch } from "vue";
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
    await kafkaStore.restartTopicConsuming(props.projectID, consumingStrategy.value, fromTime.value, offsetValue.value);
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
    field: "timestampFormatted",
    sortable: true,
  },
  { name: "partition", align: "left", label: "Partition", field: "partitionID" },
  { name: "offset", align: "left", label: "Offset", field: "offset" },
  { name: "key", align: "left", label: "Key", field: "key" },
  { name: "data", align: "left", label: "Data", field: "data" },
  { name: "headers", align: "left", label: "Headers", field: "" },
];

const consumedMessagesTablePagination = {
  rowsPerPage: 20,
  sortBy: "timestamp",
  descending: true,
};

const consumedMessagesTableRowsPerPage = [5, 10, 20, 50, 100, 200, 500];

const currentTimeTenMinutesAgo = date.subtractFromDate(Date.now(), { minutes: 10 });

const consumingStrategy = ref("time");
const consumingStrategyGroup = ref("time");
const consumingStrategyOffset = ref("");

const fromTime = ref(date.formatDate(currentTimeTenMinutesAgo, "YYYY-MM-DD HH:mm:ss Z"));
const offsetValue = ref(0);

function setCurrentTimeMinutes(minutes) {
  const currentTimeMinutesAgo = date.subtractFromDate(Date.now(), { minutes: minutes });
  fromTime.value = date.formatDate(currentTimeMinutesAgo, "YYYY-MM-DD HH:mm:ss Z");
}

watch(
  () => consumingStrategyGroup.value,
  (group, _) => {
    if (group === "time") {
      consumingStrategy.value = "time";
      consumingStrategyOffset.value = "";
    }
  }
);

function selectOffsetNewest() {
  consumingStrategy.value = "offset_newest";

  restartTopicConsuming();
}

function selectOffsetOldest() {
  consumingStrategy.value = "offset_oldest";

  restartTopicConsuming();
}

function selectOffsetSpecific() {
  consumingStrategy.value = "offset_specific";
}

function customMessagesSorting(rows, sortBy, descending) {
  const data = [...rows];

  if (sortBy) {
    data.sort((a, b) => {
      const x = descending ? b : a;
      const y = descending ? a : b;

      if (sortBy === "timestamp") {
        return x["timestampUnix"] - y["timestampUnix"];
      }
    });
  }

  return data;
}

try {
  kafkaStore.startTopicConsuming(
    props.projectID,
    consumingStrategy.value,
    currentConsumedTopic.value,
    fromTime.value,
    offsetValue.value
  );
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
        v-model="consumingStrategyGroup"
        size="sm"
        toggle-color="primary"
        :options="[
          { label: 'Time', value: 'time' },
          { label: 'Offset', value: 'offset' },
        ]"
      />

      <div v-if="consumingStrategyGroup === 'time'">
        <div class="row no-wrap">
          <q-btn-dropdown outline color="info" :label="fromTime" dropdown-icon="none" @hide="restartTopicConsuming()">
            <div class="row items-start">
              <q-date v-model="fromTime" mask="YYYY-MM-DD HH:mm:ss Z" color="primary" first-day-of-week="1" />
              <q-time v-model="fromTime" mask="YYYY-MM-DD HH:mm:ss Z" color="primary" with-seconds format24h />
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

      <div v-if="consumingStrategyGroup === 'offset'">
        <q-list>
          <q-item>
            <q-item-section avatar>
              <q-radio
                v-model="consumingStrategyOffset"
                val="newest"
                label="Newest"
                dense
                @click="selectOffsetNewest()"
              />
            </q-item-section>
          </q-item>

          <q-item>
            <q-item-section avatar>
              <q-radio
                v-model="consumingStrategyOffset"
                val="oldest"
                label="Oldest"
                dense
                @click="selectOffsetOldest()"
              />
            </q-item-section>
          </q-item>

          <q-item>
            <q-item-section avatar>
              <q-radio
                v-model="consumingStrategyOffset"
                val="specific"
                label="From offset"
                dense
                @click="selectOffsetSpecific()"
              />
            </q-item-section>

            <q-item-section>
              <q-input
                v-model="offsetValue"
                label=""
                dense
                input-style="padding-top: 0;"
                :disable="consumingStrategyOffset !== 'specific'"
              />
            </q-item-section>

            <q-item-section>
              <q-btn
                label="Load"
                color="secondary"
                @click="restartTopicConsuming()"
                :disable="consumingStrategyOffset !== 'specific'"
                size="sm"
              />
            </q-item-section>
          </q-item>
        </q-list>
      </div>

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
    </template>

    <template v-slot:after>
      <q-table
        :rows="consumedTopicMessages"
        :columns="consumedMessagesTableColumns"
        :row-key="consumedMessagesTableRowKey"
        :pagination="consumedMessagesTablePagination"
        :rows-per-page-options="consumedMessagesTableRowsPerPage"
        :loading="!consumedTopic.topicName"
        :sort-method="customMessagesSorting"
        binary-state-sort
      />
    </template>
  </q-splitter>
</template>

<style></style>
