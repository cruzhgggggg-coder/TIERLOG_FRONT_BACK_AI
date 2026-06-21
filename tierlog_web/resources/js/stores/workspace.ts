import { defineStore } from 'pinia';
import { ref } from 'vue';

export interface PanelState {
  id: string; // 'roster' | 'history' | 'feedback' | 'chat' | 'queue'
  title: string;
  x: number;
  y: number;
  w: number;
  h: number;
  visible: boolean;
  isMaximized: boolean;
  zIndex: number;
}

export const useWorkspaceStore = defineStore('workspace', () => {
  const maxZIndex = ref(10);
  const selectedStudentId = ref<number | null>(null);
  const activeLogId = ref<number | null>(null);

  const panels = ref<PanelState[]>([
    {
      id: 'roster',
      title: 'Student Roster',
      x: 20,
      y: 20,
      w: 300,
      h: 580,
      visible: true,
      isMaximized: false,
      zIndex: 1,
    },
    {
      id: 'history',
      title: 'Session History',
      x: 340,
      y: 20,
      w: 380,
      h: 580,
      visible: true,
      isMaximized: false,
      zIndex: 2,
    },
    {
      id: 'feedback',
      title: 'Feedback Composer',
      x: 740,
      y: 20,
      w: 420,
      h: 320,
      visible: true,
      isMaximized: false,
      zIndex: 3,
    },
    {
      id: 'chat',
      title: 'Advisor Chat',
      x: 1180,
      y: 20,
      w: 380,
      h: 580,
      visible: true,
      isMaximized: false,
      zIndex: 4,
    },
    {
      id: 'queue',
      title: 'Validation Queue',
      x: 740,
      y: 360,
      w: 420,
      h: 240,
      visible: false,
      isMaximized: false,
      zIndex: 5,
    },
  ]);

  function togglePanel(id: string) {
    const panel = panels.value.find((p) => p.id === id);
    if (panel) {
      panel.visible = !panel.visible;
      if (panel.visible) {
        raisePanel(id);
      }
    }
  }

  function updatePanelPosition(id: string, x: number, y: number) {
    const panel = panels.value.find((p) => p.id === id);
    if (panel) {
      // Boundaries check will be handled in visual component, but update here
      panel.x = x;
      panel.y = y;
    }
  }

  function updatePanelSize(id: string, w: number, h: number) {
    const panel = panels.value.find((p) => p.id === id);
    if (panel) {
      panel.w = Math.max(200, w);
      panel.h = Math.max(150, h);
    }
  }

  function maximizePanel(id: string) {
    const panel = panels.value.find((p) => p.id === id);
    if (panel) {
      panel.isMaximized = !panel.isMaximized;
      if (panel.isMaximized) {
        raisePanel(id);
      }
    }
  }

  function raisePanel(id: string) {
    const panel = panels.value.find((p) => p.id === id);
    if (panel) {
      maxZIndex.value++;
      panel.zIndex = maxZIndex.value;
    }
  }

  function tilePanels(canvasWidth: number, canvasHeight: number) {
    const visiblePanels = panels.value.filter((p) => p.visible);
    if (visiblePanels.length === 0) return;

    const count = visiblePanels.length;
    const padding = 15;
    const availableWidth = canvasWidth - padding * (count + 1);
    const width = Math.floor(availableWidth / count);
    const height = canvasHeight - padding * 2;

    visiblePanels.forEach((panel, index) => {
      panel.isMaximized = false;
      panel.x = padding + index * (width + padding);
      panel.y = padding;
      panel.w = width;
      panel.h = height;
      raisePanel(panel.id);
    });
  }

  function setSelectedStudentId(id: number | null) {
    selectedStudentId.value = id;
    if (id === null) {
      activeLogId.value = null;
    }
  }

  return {
    panels,
    maxZIndex,
    selectedStudentId,
    activeLogId,
    togglePanel,
    updatePanelPosition,
    updatePanelSize,
    maximizePanel,
    raisePanel,
    tilePanels,
    setSelectedStudentId,
  };
});
