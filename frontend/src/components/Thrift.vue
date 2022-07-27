<script>
import { defineComponent } from "vue";
import { mapState } from "pinia";

import { useThriftStore } from "../stores/thrift";
import ThriftForm from "./ThriftForm.vue";

export default defineComponent({
  name: "Thrift",
  props: {
    projectID: String,
  },
  components: { ThriftForm },
  beforeCreate() {
    useThriftStore().loadState();
  },
  computed: {
    ...mapState(useThriftStore, ["projects"]),
    importPathList() {
      if (useThriftStore().projects[this.projectID]) {
        return useThriftStore().projects[this.projectID].importPathList;
      }
    },
    nodes() {
      if (useThriftStore().projects[this.projectID]) {
        return useThriftStore().projects[this.projectID].nodes;
      }
    },
    forms() {
      if (useThriftStore().projects[this.projectID]) {
        return useThriftStore().projects[this.projectID].forms;
      }
    },
    formIDs() {
      if (useThriftStore().projects[this.projectID]) {
        return useThriftStore().projects[this.projectID].formIDs;
      }
    },
    currentFormID: {
      get() {
        if (useThriftStore().projects[this.projectID]) {
          return useThriftStore().projects[this.projectID].currentFormID;
        }
      },
      async set(currentFormID) {
        await useThriftStore().saveCurrentFormID(this.projectID, currentFormID);
      },
    },
    splitterWidth: {
      get() {
        if (useThriftStore().projects[this.projectID]) {
          return useThriftStore().projects[this.projectID].splitterWidth;
        }
      },
      async set(splitterWidth) {
        await useThriftStore().saveSplitterWidth(this.projectID, splitterWidth);
      },
    },
    selectedFunction: {
      get() {
        if (useThriftStore().projects[this.projectID]) {
          const currentFormID = useThriftStore().projects[this.projectID].currentFormID;
          const currentForm = useThriftStore().projects[this.projectID].forms[currentFormID];

          if (currentForm) {
            return currentForm.selectedFunctionID;
          }
        }
      },
      async set(selectedFunctionID) {
        await useThriftStore().selectFunction(
          this.projectID,
          this.projects[this.projectID].currentFormID,
          selectedFunctionID
        );
      },
    },
  },
  watch: {
    currentFormID(newCurrentFormID, oldCurrentFormID) {
      if (newCurrentFormID === oldCurrentFormID) {
        return;
      }

      const formID = newCurrentFormID || oldCurrentFormID;
      const form = useThriftStore().projects[this.projectID].forms[formID];

      if (form.selectedFunctionID && this.selectedFunction !== form.selectedFunctionID) {
        this.selectedFunction = form.selectedFunctionID;
      }
    },
  },
  methods: {
    async openFilePath() {
      const store = useThriftStore();

      await store.openFilePath(this.projectID);
    },

    async createNewForm() {
      const store = useThriftStore();

      await store.createNewForm(this.projectID);
    },

    async closeFormTab(event, formID) {
      event.preventDefault();

      const store = useThriftStore();

      await store.removeForm(this.projectID, formID);
    },
  },
});
</script>

<template>
  <div class="full-height">
    <q-splitter v-if="splitterWidth" v-model="splitterWidth" class="full-height" :limits="[20, 80]">
      <template v-slot:before>
        <q-btn size="sm" label="Open Thrift file" color="primary" @click="openFilePath" />

        <q-tree
          v-if="(nodes || []).length > 0"
          ref="serviceTree"
          :nodes="nodes"
          default-expand-all
          no-selection-unset
          v-model:selected="selectedFunction"
          node-key="id"
        />
      </template>

      <template v-slot:after>
        <q-tabs v-model="currentFormID" align="left" outside-arrows mobile-arrows dense no-caps>
          <q-tab :name="formID" v-for="formID in formIDs" :key="`tab-${formID}`">
            <div class="row justify-between">
              <div class="col q-tab__label">
                <div v-if="forms[formID].selectedFunctionID.length < 15">
                  {{ forms[formID].selectedFunctionID || "New Form" }}
                </div>

                <div v-else class="thrift-form-tab-name">
                  <div class="start">{{ forms[formID].selectedFunctionID.substring(0, 20) }}</div>
                  <div class="end">
                    {{
                      forms[formID].selectedFunctionID.substring(
                        forms[formID].selectedFunctionID.length - 20 > 20
                          ? forms[formID].selectedFunctionID.length - 20
                          : 20
                      )
                    }}
                  </div>
                </div>
              </div>

              <div class="col-1">
                <q-btn
                  class="inline"
                  icon="close"
                  size="10px"
                  style="width: 20px"
                  flat
                  rounded
                  dense
                  :disable="Object.keys(this.forms).length === 1"
                  @click="closeFormTab($event, formID)"
                />
              </div>
            </div>
          </q-tab>

          <q-btn @click="createNewForm" icon="add" color="primary" />
        </q-tabs>

        <q-separator />

        <q-tab-panels id="formContainer" v-model="currentFormID" animated>
          <q-tab-panel :name="formID" v-for="(form, formID) in forms" :key="`tab-panel-${formID}`">
            <ThriftForm :formID="formID" :projectID="this.projectID" :selectedFunctionID="this.selectedFunction" />
          </q-tab-panel>
        </q-tab-panels>
      </template>
    </q-splitter>
  </div>
</template>

<style>
#formContainer {
  height: calc(100% - 48px) !important;
}

.thrift-form-tab-name {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  justify-content: flex-start;
}

.thrift-form-tab-name > .start {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex-shrink: 1;
}

.thrift-form-tab-name > .end {
  white-space: nowrap;
  flex-basis: content;
  flex-grow: 0;
  flex-shrink: 0;
}

.q-tree__node--selected .q-tree__node-header-content {
  color: #3498db;
}

.q-tabs__content--align-center .q-tab {
  flex: 1 1 auto;
}
</style>
