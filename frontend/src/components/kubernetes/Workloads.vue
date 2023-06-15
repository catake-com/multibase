<script setup>
import { computed, ref, watch } from "vue";
import { useKubernetesStore } from "../../stores/kubernetes";
import { useQuasar } from "quasar";

const quasar = useQuasar();

const props = defineProps({
  projectID: String,
});

const tablePodsLoading = ref(true);

const kubernetesStore = useKubernetesStore();
const currentWorkloadTab = ref("overview");

const isPortForwarded = computed(() => kubernetesStore.projectState(props.projectID).isPortForwarded);
const namespaces = computed(() => kubernetesStore.namespaces(props.projectID));
const workloadsPodsData = computed(() => kubernetesStore.workloadsPodsData(props.projectID));

const selectedNamespace = computed({
  get() {
    return kubernetesStore.projectState(props.projectID).selectedNamespace;
  },
  async set(selectedNamespace) {
    tablePodsLoading.value = true;

    await kubernetesStore.selectNamespace(props.projectID, selectedNamespace);
    await kubernetesStore.loadWorkloadsPodsData(props.projectID);

    tablePodsLoading.value = false;
  },
});

async function fetchData(tab) {
  switch (tab) {
    case "pods":
      tablePodsLoading.value = true;

      await Promise.all([
        kubernetesStore.loadNamespaces(props.projectID),
        kubernetesStore.loadWorkloadsPodsData(props.projectID),
      ]);

      tablePodsLoading.value = false;

      break;
  }
}

watch(
  () => currentWorkloadTab.value,
  async (newTab, oldTab) => {
    if (newTab === oldTab) {
      return;
    }

    await fetchData(newTab);
  }
);

const tablePodsFilter = ref("");
const tablePodsColumns = [
  {
    name: "name",
    label: "Name",
    align: "left",
    field: "name",
  },
  { name: "namespace", align: "left", label: "Namespace", field: "namespace" },
  { name: "ports", align: "left", label: "Ports", field: "ports" },
];

const tablePodsRowsPerPage = [5, 10, 20, 50, 100, 200, 500];

const tablePodsPagination = {
  rowsPerPage: 50,
};

function tablePodsFilterMethod(rows, query, cols, getCellValue) {
  if (query === "") {
    return rows;
  }

  const queryLowerCase = query.toLowerCase();

  return rows.filter((row) => row.name.toLowerCase().includes(queryLowerCase));
}

async function startPortForwarding(namespace, pod, ports) {
  try {
    await kubernetesStore.startPortForwarding(props.projectID, namespace, pod, ports);

    quasar.notify({ type: "positive", message: "Port Forwarding started" });
  } catch (error) {
    quasar.notify({ type: "negative", message: error });
  }
}

async function stopPortForwarding() {
  try {
    await kubernetesStore.stopPortForwarding(props.projectID);

    quasar.notify({ type: "positive", message: "Port Forwarding stopped" });
  } catch (error) {
    quasar.notify({ type: "negative", message: error });
  }
}
</script>

<template>
  <div class="full-height">
    <q-tabs v-model="currentWorkloadTab">
      <q-tab name="overview" label="OVERVIEW" />
      <q-tab name="pods" label="PODS" />
      <q-tab name="deployments" label="DEPLOYMENTS" />
    </q-tabs>

    <q-tab-panels v-model="currentWorkloadTab">
      <q-tab-panel name="overview"> Overview </q-tab-panel>

      <q-tab-panel name="pods">
        <q-btn
          v-if="isPortForwarded"
          size="sm"
          label="Stop Port Forwarding"
          color="secondary"
          @click="stopPortForwarding()"
        />

        <q-table
          :filter="tablePodsFilter"
          :filter-method="tablePodsFilterMethod"
          :rows="workloadsPodsData.pods"
          :columns="tablePodsColumns"
          row-key="name"
          :pagination="tablePodsPagination"
          :rows-per-page-options="tablePodsRowsPerPage"
          :loading="tablePodsLoading"
        >
          <template v-slot:top-right>
            <q-select
              dense
              clearable
              v-model="selectedNamespace"
              :options="namespaces"
              label="Namespace"
              style="width: 200px; margin-right: 20px"
            />

            <q-input borderless dense v-model="tablePodsFilter" placeholder="Search Pods...">
              <template v-slot:append>
                <q-icon name="search" />
              </template>
            </q-input>
          </template>

          <template v-slot:body-cell-ports="props">
            <q-td :props="props">
              <div v-for="port in props.row.ports" :key="`${port.name}-${port.containerPort}`">
                <span>{{ port.name }}</span>
                <span>:</span>
                <span>{{ `${port.containerPort}:${port.containerPort}` }}</span>
                <q-btn
                  v-if="!isPortForwarded"
                  size="sm"
                  label="Forward"
                  color="secondary"
                  @click="
                    startPortForwarding(
                      props.row.namespace,
                      props.row.name,
                      `${port.containerPort}:${port.containerPort}`
                    )
                  "
                />
              </div>
            </q-td>
          </template>
        </q-table>
      </q-tab-panel>

      <q-tab-panel name="deployments"> Deployments </q-tab-panel>
    </q-tab-panels>
  </div>
</template>

<style></style>
