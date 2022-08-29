<script>
import { defineComponent } from "vue";
import { mapState } from "pinia";

import { useKafkaStore } from "../stores/kafka";

export default defineComponent({
  name: "Kafka",
  props: {
    projectID: String,
  },
  data() {
    return {
      splitterWidth: 12,
    };
  },
  beforeCreate() {
    useKafkaStore().loadState();
  },
  computed: {
    ...mapState(useKafkaStore, ["projects"]),
    currentTab: {
      get() {
        if (useKafkaStore().projects[this.projectID]) {
          return useKafkaStore().projects[this.projectID].currentTab;
        }
      },
      async set(currentTab) {
        await useKafkaStore().saveCurrentTab(this.projectID, currentTab);
      },
    },
    address: {
      get() {
        if (useKafkaStore().projects[this.projectID]) {
          return useKafkaStore().projects[this.projectID].address;
        }
      },
      async set(address) {
        await useKafkaStore().saveAddress(this.projectID, address);
      },
    },
    authMethod: {
      get() {
        if (useKafkaStore().projects[this.projectID]) {
          return useKafkaStore().projects[this.projectID].authMethod;
        }
      },
      async set(authMethod) {
        await useKafkaStore().saveAuthMethod(this.projectID, authMethod);
      },
    },
    authUsername: {
      get() {
        if (useKafkaStore().projects[this.projectID]) {
          return useKafkaStore().projects[this.projectID].authUsername;
        }
      },
      async set(authUsername) {
        await useKafkaStore().saveAuthUsername(this.projectID, authUsername);
      },
    },
    authPassword: {
      get() {
        if (useKafkaStore().projects[this.projectID]) {
          return useKafkaStore().projects[this.projectID].authPassword;
        }
      },
      async set(authPassword) {
        await useKafkaStore().saveAuthPassword(this.projectID, authPassword);
      },
    },
  },
  methods: {
    async connect() {},
  },
});
</script>

<template>
  <div class="full-height">
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
                <q-input v-model="authPassword" label="Password" debounce="500" />
              </div>

              <q-btn label="Connect" color="secondary" @click="connect" />
            </q-form>
          </q-tab-panel>

          <q-tab-panel name="brokers"> brokers </q-tab-panel>

          <q-tab-panel name="topics">
            <q-markup-table>
              <thead>
                <tr>
                  <th class="text-left">Dessert (100g serving)</th>
                  <th class="text-right">Calories</th>
                  <th class="text-right">Fat (g)</th>
                  <th class="text-right">Carbs (g)</th>
                  <th class="text-right">Protein (g)</th>
                  <th class="text-right">Sodium (mg)</th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td class="text-left">Frozen Yogurt</td>
                  <td class="text-right">159</td>
                  <td class="text-right">6</td>
                  <td class="text-right">24</td>
                  <td class="text-right">4</td>
                  <td class="text-right">87</td>
                </tr>
                <tr>
                  <td class="text-left">Ice cream sandwich</td>
                  <td class="text-right">237</td>
                  <td class="text-right">9</td>
                  <td class="text-right">37</td>
                  <td class="text-right">4.3</td>
                  <td class="text-right">129</td>
                </tr>
                <tr>
                  <td class="text-left">Eclair</td>
                  <td class="text-right">262</td>
                  <td class="text-right">16</td>
                  <td class="text-right">23</td>
                  <td class="text-right">6</td>
                  <td class="text-right">337</td>
                </tr>
                <tr>
                  <td class="text-left">Cupcake</td>
                  <td class="text-right">305</td>
                  <td class="text-right">3.7</td>
                  <td class="text-right">67</td>
                  <td class="text-right">4.3</td>
                  <td class="text-right">413</td>
                </tr>
                <tr>
                  <td class="text-left">Gingerbread</td>
                  <td class="text-right">356</td>
                  <td class="text-right">16</td>
                  <td class="text-right">49</td>
                  <td class="text-right">3.9</td>
                  <td class="text-right">327</td>
                </tr>
              </tbody>
            </q-markup-table>
          </q-tab-panel>

          <q-tab-panel name="consumers"> </q-tab-panel>
        </q-tab-panels>
      </template>
    </q-splitter>
  </div>
</template>

<style></style>
