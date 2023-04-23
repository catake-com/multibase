<script setup>
import { computed } from "vue";
import { useQuasar } from "quasar";
import { useKafkaStore } from "../stores/kafka";

const quasar = useQuasar();

const kafkaStore = useKafkaStore();

const props = defineProps({
  projectID: String,
});

const splitterWidth = 12;
const splitterWidthConsuming = 12;

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

const hoursAgo = computed({
  get() {
    return kafkaStore.consumingSession(props.projectID).hoursAgo;
  },
  async set(hoursAgo) {
    kafkaStore.consumingSession(props.projectID).hoursAgo = hoursAgo;
  },
});

const currentTopic = computed(() => kafkaStore.consumingSession(props.projectID).currentTopic);
const topics = computed(() => kafkaStore.topicsData(props.projectID));
const brokers = computed(() => kafkaStore.brokersData(props.projectID));
const consumers = computed(() => kafkaStore.consumersData(props.projectID));
const consumedTopic = computed(() => kafkaStore.consumedTopic(props.projectID));
const consumedTopicMessages = computed(() => kafkaStore.consumedTopicMessages(props.projectID));

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

async function startTopicConsuming(topic) {
  try {
    await kafkaStore.startTopicConsuming(props.projectID, topic, 1);
    quasar.notify({ type: "positive", message: "Started consuming" });
  } catch (error) {
    quasar.notify({ type: "negative", message: error });
  }
}

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
  <div class="full-height">
    <div v-if="currentTopic">
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
    </div>

    <div v-else>
      <q-splitter v-model="splitterWidth" class="full-height" disable>
        <template v-slot:before>
          <q-tabs v-model="currentTab" vertical>
            <q-tab name="overview" icon="home" label="Overview" />
            <q-tab name="brokers" icon="lan" label="Brokers" />
            <q-tab name="topics" icon="storage" label="Topics" />
            <q-tab name="consumers" icon="browser_updated" label="Consumers" />
          </q-tabs>
        </template>

        <template v-slot:after>
          <q-tab-panels v-model="currentTab" animated vertical>
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
                <q-markup-table>
                  <thead>
                    <tr>
                      <th class="text-left">Topic Name</th>
                      <th class="text-left">Partitions</th>
                      <th class="text-left">Count</th>
                      <th class="text-left">Actions</th>
                    </tr>
                  </thead>

                  <tbody>
                    <tr v-for="topic in topics.list" :key="topic.name">
                      <td class="text-left">{{ topic.name }}</td>
                      <td class="text-left">{{ topic.partitionCount }}</td>
                      <td class="text-left">{{ topic.messageCount }}</td>
                      <td class="text-left">
                        <q-btn label="Consume" color="secondary" @click="startTopicConsuming(topic.name)" />
                      </td>
                    </tr>
                  </tbody>
                </q-markup-table>
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
