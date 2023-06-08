<script setup>
import { computed, ref } from "vue";
import { useQuasar } from "quasar";
import { useKafkaStore } from "../stores/kafka";
import Consume from "./kafka/Consume.vue";

const quasar = useQuasar();

const kafkaStore = useKafkaStore();

const props = defineProps({
  projectID: String,
});

const splitterWidth = 12;

await kafkaStore.loadProject(props.projectID);

const currentTab = computed({
  get() {
    return kafkaStore.projectState(props.projectID).currentTab;
  },
  async set(currentTab) {
    const projectState = kafkaStore.projectState(props.projectID);
    projectState.currentTab = currentTab;

    await kafkaStore.saveState(props.projectID, projectState);
  },
});

const address = computed({
  get() {
    return kafkaStore.projectState(props.projectID).address;
  },
  async set(address) {
    const projectState = kafkaStore.projectState(props.projectID);
    projectState.address = address;

    await kafkaStore.saveState(props.projectID, projectState);
  },
});

const authMethod = computed({
  get() {
    return kafkaStore.projectState(props.projectID).authMethod;
  },
  async set(authMethod) {
    const projectState = kafkaStore.projectState(props.projectID);
    projectState.authMethod = authMethod;

    await kafkaStore.saveState(props.projectID, projectState);
  },
});

const authUsername = computed({
  get() {
    return kafkaStore.projectState(props.projectID).authUsername;
  },
  async set(authUsername) {
    const projectState = kafkaStore.projectState(props.projectID);
    projectState.authUsername = authUsername;

    await kafkaStore.saveState(props.projectID, projectState);
  },
});

const authPassword = computed({
  get() {
    return kafkaStore.projectState(props.projectID).authPassword;
  },
  async set(authPassword) {
    const projectState = kafkaStore.projectState(props.projectID);
    projectState.authPassword = authPassword;

    await kafkaStore.saveState(props.projectID, projectState);
  },
});

const currentConsumedTopic = computed(() => kafkaStore.initiatedTopicConsuming(props.projectID).topicName);
const topics = computed(() => kafkaStore.topicsData(props.projectID));
const brokers = computed(() => kafkaStore.brokersData(props.projectID));
const consumers = computed(() => kafkaStore.consumersData(props.projectID));

const topicTableColumns = [
  {
    name: "name",
    label: "Topic Name",
    align: "left",
    field: "name",
  },
  { name: "partitionCount", align: "left", label: "Partitions", field: "partitionCount" },
  { name: "messageCount", align: "left", label: "Count", field: "messageCount" },
  { name: "actions", align: "left", label: "Actions", field: "name" },
];

const topicTableRowsPerPage = [5, 10, 20, 50, 100, 200, 500];

const topicTablePagination = {
  rowsPerPage: 100,
};

const topicTableFilter = ref("");

function topicTableFilterMethod(rows, query, cols, getCellValue) {
  if (query === "") {
    return rows;
  }

  const queryLowerCase = query.toLowerCase();

  return rows.filter((row) => row.name.toLowerCase().includes(queryLowerCase));
}

async function connect() {
  try {
    await kafkaStore.connect(props.projectID);
    await Promise.all([
      kafkaStore.loadTopics(props.projectID),
      kafkaStore.loadBrokers(props.projectID),
      kafkaStore.loadConsumers(props.projectID),
    ]);

    quasar.notify({ type: "positive", message: "Connected" });
  } catch (error) {
    quasar.notify({ type: "negative", message: error });
  }
}

function initiateTopicConsuming(topic) {
  kafkaStore.initiateTopicConsuming(props.projectID, topic);
}
</script>

<template>
  <div class="full-height">
    <div v-if="currentConsumedTopic">
      <Consume :projectID="projectID"></Consume>
    </div>

    <div v-else>
      <q-splitter v-model="splitterWidth" class="full-height" disable>
        <template v-slot:before>
          <q-tabs v-model="currentTab" vertical>
            <q-tab name="overview" icon="home" label="Overview" />
            <q-tab name="brokers" icon="lan" label="Brokers" :disable="!brokers.isConnected" />
            <q-tab name="topics" icon="storage" label="Topics" :disable="!topics.isConnected" />
            <q-tab name="consumers" icon="browser_updated" label="Consumers" :disable="!consumers.isConnected" />
          </q-tabs>
        </template>

        <template v-slot:after>
          <q-tab-panels v-model="currentTab" vertical>
            <q-tab-panel name="overview">
              <q-form class="q-gutter-md full-height">
                <q-input v-model="address" label="Address" debounce="500" />

                <div>
                  <q-radio v-model="authMethod" val="plaintext" label="Plaintext" dense />
                  <q-radio v-model="authMethod" val="sasl_ssl" label="SASL SSL" dense />
                </div>

                <div v-if="authMethod === 'sasl_ssl'">
                  <q-input v-model="authUsername" label="Username" debounce="500" />
                  <q-input v-model="authPassword" label="Password" debounce="500" type="password" />
                </div>

                <q-btn label="Connect" color="secondary" @click="connect" />
              </q-form>
            </q-tab-panel>

            <q-tab-panel name="brokers">
              <div v-if="brokers.isConnected">
                <q-markup-table>
                  <thead>
                    <tr>
                      <th class="text-left">ID</th>
                      <th class="text-left">Rack</th>
                      <th class="text-left">Listener</th>
                      <th class="text-left">Actions</th>
                    </tr>
                  </thead>

                  <tbody>
                    <tr v-for="broker in brokers.list" :key="broker.id">
                      <td class="text-left">{{ broker.id }}</td>
                      <td class="text-left">{{ broker.rack }}</td>
                      <td class="text-left">{{ `${broker.host}:${broker.port}` }}</td>
                      <td class="text-left"></td>
                    </tr>
                  </tbody>
                </q-markup-table>
              </div>

              <div v-else>Not connected to Kafka</div>
            </q-tab-panel>

            <q-tab-panel name="topics">
              <div v-if="topics.isConnected">
                <q-table
                  :filter="topicTableFilter"
                  :filter-method="topicTableFilterMethod"
                  :rows="topics.list"
                  :columns="topicTableColumns"
                  row-key="name"
                  :pagination="topicTablePagination"
                  :rows-per-page-options="topicTableRowsPerPage"
                >
                  <template v-slot:header-cell-name="props">
                    <q-th :props="props">
                      <div class="row items-center">
                        <div style="margin-right: 10px">Topic Name</div>

                        <q-input borderless dense v-model="topicTableFilter" placeholder="Filter" style="width: 50%">
                        </q-input>
                      </div>
                    </q-th>
                  </template>

                  <template v-slot:body-cell-actions="props">
                    <q-td :props="props">
                      <q-btn label="Consume" color="secondary" @click="initiateTopicConsuming(props.value)" />
                    </q-td>
                  </template>
                </q-table>
              </div>

              <div v-else>Not connected to Kafka</div>
            </q-tab-panel>

            <q-tab-panel name="consumers">
              <div v-if="consumers.isConnected">
                <q-markup-table>
                  <thead>
                    <tr>
                      <th class="text-left">Name</th>
                      <th class="text-left">State</th>
                      <th class="text-left">Actions</th>
                    </tr>
                  </thead>

                  <tbody>
                    <tr v-for="consumer in consumers.list" :key="consumer.id">
                      <td class="text-left">{{ consumer.name }}</td>
                      <td class="text-left">{{ consumer.state }}</td>
                      <td class="text-left"></td>
                    </tr>
                  </tbody>
                </q-markup-table>
              </div>

              <div v-else>Not connected to Kafka</div>
            </q-tab-panel>
          </q-tab-panels>
        </template>
      </q-splitter>
    </div>
  </div>
</template>

<style></style>
