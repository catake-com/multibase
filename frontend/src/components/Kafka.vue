<script>
import { defineComponent } from "vue";
import { useKafkaStore } from "../stores/kafka";

export default defineComponent({
  name: "Kafka",
  props: {
    projectID: String,
  },
  data() {
    return {
      splitterWidth: 12,
      splitterWidthConsuming: 12,
    };
  },
  beforeCreate() {
    useKafkaStore().loadState();
  },
  computed: {
    currentTab: {
      get() {
        if (useKafkaStore().main.projects[this.projectID]) {
          return useKafkaStore().main.projects[this.projectID].currentTab;
        }
      },
      async set(currentTab) {
        await useKafkaStore().saveCurrentTab(this.projectID, currentTab);
      },
    },
    address: {
      get() {
        if (useKafkaStore().main.projects[this.projectID]) {
          return useKafkaStore().main.projects[this.projectID].address;
        }
      },
      async set(address) {
        await useKafkaStore().saveAddress(this.projectID, address);
      },
    },
    authMethod: {
      get() {
        if (useKafkaStore().main.projects[this.projectID]) {
          return useKafkaStore().main.projects[this.projectID].authMethod;
        }
      },
      async set(authMethod) {
        await useKafkaStore().saveAuthMethod(this.projectID, authMethod);
      },
    },
    authUsername: {
      get() {
        if (useKafkaStore().main.projects[this.projectID]) {
          return useKafkaStore().main.projects[this.projectID].authUsername;
        }
      },
      async set(authUsername) {
        await useKafkaStore().saveAuthUsername(this.projectID, authUsername);
      },
    },
    authPassword: {
      get() {
        if (useKafkaStore().main.projects[this.projectID]) {
          return useKafkaStore().main.projects[this.projectID].authPassword;
        }
      },
      async set(authPassword) {
        await useKafkaStore().saveAuthPassword(this.projectID, authPassword);
      },
    },
    hoursAgo: {
      get() {
        if (useKafkaStore().session[this.projectID]) {
          return useKafkaStore().session[this.projectID].hoursAgo;
        }
      },
      set(hoursAgo) {
        useKafkaStore().session[this.projectID].hoursAgo = parseInt(hoursAgo);
      },
    },
    currentTopic() {
      if (useKafkaStore().session[this.projectID]) {
        return useKafkaStore().session[this.projectID].currentTopic;
      }
    },
    topics() {
      return useKafkaStore().topics[this.projectID];
    },
    brokers() {
      return useKafkaStore().brokers[this.projectID];
    },
    consumers() {
      return useKafkaStore().consumers[this.projectID];
    },
    consumedTopic() {
      return useKafkaStore().consumedTopic[this.projectID];
    },
    consumedTopicMessages() {
      return useKafkaStore().consumedTopicMessages[this.projectID];
    },
  },
  methods: {
    async connect() {
      try {
        await useKafkaStore().connect(this.projectID);
        await useKafkaStore().loadTopics(this.projectID);
        await useKafkaStore().loadBrokers(this.projectID);
        await useKafkaStore().loadConsumers(this.projectID);
        this.$q.notify({ type: "positive", message: "Connected" });
      } catch (error) {
        this.$q.notify({ type: "negative", message: error });
      }
    },

    async startTopicConsuming(topic) {
      try {
        await useKafkaStore().startTopicConsuming(this.projectID, topic, 1);
        this.$q.notify({ type: "positive", message: "Started consuming" });
      } catch (error) {
        this.$q.notify({ type: "negative", message: error });
      }
    },

    async stopTopicConsuming() {
      try {
        await useKafkaStore().stopTopicConsuming(this.projectID);
        this.$q.notify({ type: "positive", message: "Stopped consuming" });
      } catch (error) {
        this.$q.notify({ type: "negative", message: error });
      }
    },

    async restartTopicConsuming() {
      try {
        await useKafkaStore().restartTopicConsuming(this.projectID);
        this.$q.notify({ type: "positive", message: "Restarted consuming" });
      } catch (error) {
        this.$q.notify({ type: "negative", message: error });
      }
    },
  },
});
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
          <q-markup-table>
            <thead>
              <tr>
                <th class="text-left">Timestamp</th>
                <th class="text-left">Partition</th>
                <th class="text-left">Offset</th>
                <th class="text-left">Key</th>
                <th class="text-left">Data</th>
                <th class="text-left">Headers</th>
              </tr>
            </thead>

            <tbody>
              <tr v-for="message in consumedTopicMessages" :key="`${message.partitionID}_${message.offset}`">
                <td class="text-left">{{ message.timestamp }}</td>
                <td class="text-left">{{ message.partitionID }}</td>
                <td class="text-left">{{ message.offset }}</td>
                <td class="text-left">{{ message.key }}</td>
                <td class="text-left">{{ message.data }}</td>
                <td class="text-left"></td>
              </tr>
            </tbody>
          </q-markup-table>
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
                  <q-radio v-model="authMethod" val="saslssl" label="SASL SSL" dense />
                </div>

                <div v-if="authMethod === 'saslssl'">
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
