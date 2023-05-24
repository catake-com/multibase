<script setup>
import { computed } from "vue";
import { useQuasar } from "quasar";
import { useKubernetesStore } from "../stores/kubernetes";

const quasar = useQuasar();

const kubernetesStore = useKubernetesStore();

const props = defineProps({
  projectID: String,
});

const splitterWidth = 12;

await kubernetesStore.loadProject(props.projectID);
await kubernetesStore.loadOverviewData(props.projectID);

const currentTab = computed({
  get() {
    return kubernetesStore.projectState(props.projectID).currentTab;
  },
  async set(currentTab) {
    await kubernetesStore.saveCurrentTab(props.projectID, currentTab);
  },
});

const isConnected = computed(() => kubernetesStore.projectState(props.projectID).isConnected);
const selectedContext = computed(() => kubernetesStore.projectState(props.projectID).selectedContext);
const selectedNamespace = computed(() => kubernetesStore.projectState(props.projectID).selectedNamespace);
const overviewData = computed(() => kubernetesStore.overviewData(props.projectID));

async function connect(selectedCluster) {
  try {
    await kubernetesStore.connect(props.projectID, selectedCluster);

    quasar.notify({ type: "positive", message: "Connected" });
  } catch (error) {
    quasar.notify({ type: "negative", message: error });
  }
}
</script>

<template>
  <div class="full-height">
    <q-splitter v-model="splitterWidth" class="full-height" disable>
      <template v-slot:before>
        <q-tabs v-model="currentTab" vertical>
          <q-tab name="overview" icon="home" label="Overview" />
          <q-tab name="workloads" icon="dvr" label="Workloads" :disable="!isConnected" />
        </q-tabs>
      </template>

      <template v-slot:after>
        <q-tab-panels v-model="currentTab" animated vertical>
          <q-tab-panel name="overview">
            <q-markup-table style="margin-bottom: 25px">
              <tbody>
                <tr>
                  <td class="text-left">Context</td>
                  <td class="text-left">
                    <div v-if="isConnected">
                      {{ selectedContext }}
                    </div>

                    <div v-else>...</div>
                  </td>
                </tr>

                <tr>
                  <td class="text-left">Namespace</td>
                  <td class="text-left">
                    <div v-if="isConnected">
                      {{ selectedNamespace }}
                    </div>

                    <div v-else>...</div>
                  </td>
                </tr>
              </tbody>
            </q-markup-table>

            <q-markup-table>
              <thead>
                <tr>
                  <th class="text-left">Name</th>
                  <th class="text-left">Cluster</th>
                  <th class="text-left"></th>
                </tr>
              </thead>

              <tbody>
                <tr v-for="context in overviewData.contexts" :key="context.name">
                  <td class="text-left">{{ context.name }}</td>
                  <td class="text-left">{{ context.cluster }}</td>
                  <td class="text-left">
                    <q-btn v-if="!isConnected" label="Connect" color="secondary" @click="connect(context.name)" />
                  </td>
                </tr>
              </tbody>
            </q-markup-table>
          </q-tab-panel>

          <q-tab-panel name="workloads">
            <div v-if="isConnected"></div>

            <div v-else>Not connected to Kubernetes</div>
          </q-tab-panel>
        </q-tab-panels>
      </template>
    </q-splitter>
  </div>
</template>

<style></style>
