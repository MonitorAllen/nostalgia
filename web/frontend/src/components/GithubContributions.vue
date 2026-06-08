<script setup lang="ts">
import { onMounted, onUnmounted, ref, nextTick } from "vue";
import { useUserStore } from "@/store/module/user";
import dayjs from "dayjs";

// @ts-ignore
import CalHeatmap from "cal-heatmap";
// @ts-ignore
import LegendLite from "cal-heatmap/plugins/LegendLite";
// @ts-ignore
import Tooltip from "cal-heatmap/plugins/Tooltip";
// @ts-ignore
import CalendarLabel from "cal-heatmap/plugins/CalendarLabel";

import "cal-heatmap/cal-heatmap.css";

const userStore = useUserStore();

const heatmap: CalHeatmap = new CalHeatmap();

// 容器 ref，用来获取实际宽度
const heatmapContainer = ref<HTMLElement | null>(null);
const isUnavailable = ref(false);

// 缩放函数：根据容器宽度等比缩小 SVG，避免超出
const resizeHeatmap = () => {
  const container = heatmapContainer.value;
  if (!container) return;

  const svg = container.querySelector("svg") as SVGSVGElement | null;
  if (!svg) return;

  const containerWidth = container.clientWidth;

  // 优先使用 viewBox 的宽高，没有的话再退回到 getBBox / clientWidth
  const vb = svg.viewBox.baseVal;
  const svgWidth =
      (vb && vb.width) || svg.getBBox().width || svg.clientWidth || containerWidth;
  const svgHeight =
      (vb && vb.height) || svg.getBBox().height || svg.clientHeight || 0;

  if (!svgWidth || !svgHeight) return;

  // 只在容器比 SVG 小时缩放，宽屏时保持 1
  const scale = containerWidth < svgWidth ? containerWidth / svgWidth : 1;

  svg.style.transformOrigin = "left top";
  svg.style.transform = `scale(${scale})`;

  // 让容器高度匹配缩放后的 SVG，避免被裁剪
  container.style.height = `${svgHeight * scale}px`;
};

let resizeTimer: number | undefined;
const onResize = () => {
  if (resizeTimer) window.clearTimeout(resizeTimer);
  // 防抖，避免频繁计算
  resizeTimer = window.setTimeout(() => {
    resizeHeatmap();
  }, 200);
};

onMounted(async () => {
  let contributions: any[] = [];

  try {
    const res: any = await userStore.contributions();
    contributions = [...(res.data?.contributions ?? [])].reverse();
  } catch {
    contributions = [];
  }

  if (contributions.length === 0) {
    isUnavailable.value = true;
    return;
  }

  await heatmap.paint(
      {
        itemSelector: "#cal-heatmap", // 显式指定你的容器
        data: {
          source: contributions,
          type: "json",
          x: "date",
          y: (d: any) => +d["intensity"],
        },
        date: {
          start: new Date(contributions[0].date),
          max: new Date(Date.now()),
        },
        range: 4,
        domain: {
          type: "month",
          gutter: 4,
          label: {
            text: "MMM",
            position: "top",
            textAlign: "start",
            offset: { y: 4 },
          },
        },
        subDomain: {
          type: "ghDay",
        },
        scale: {
          color: {
            type: "threshold",
            range: ["#ebedf0", "#9be9a8", "#40c463", "#30a14e", "#216e39"],
            domain: [1, 3, 5, 10],
          },
        },
      },
      [
        [
          Tooltip,
          {
            text: function (date: any, value: number, dayjsDate: any) {
              return value
                  ? "在 " +
                  dayjsDate.format("YYYY-MM-DD") +
                  " 有 " +
                  value +
                  " 次活动"
                  : dayjsDate.format("YYYY-MM-DD") + " 这一天很懒";
            },
          },
        ],
        [
          LegendLite,
          {
            itemSelector: "#ex-ghDay-legend",
            radius: 0,
            width: 11,
            height: 11,
            gutter: 0,
          },
        ],
        [
          CalendarLabel,
          {
            width: 30,
            textAlign: "start",
            // @ts-ignore
            text: () =>
                dayjs.weekdaysShort().map((d: any, i: any) => (i % 2 == 0 ? "" : d)),
            padding: [25, 0, 0, 0],
          },
        ],
      ]
  );

  // 等 SVG 渲染出来后再做一次缩放
  await nextTick();
  resizeHeatmap();

  window.addEventListener("resize", onResize);
});

onUnmounted(() => {
  window.removeEventListener("resize", onResize);
  heatmap.destroy();
});
</script>

<template>
  <div class="flex flex-col items-center justify-center">
    <div
        v-if="isUnavailable"
        class="grid min-h-32 w-full place-items-center rounded-lg border border-dashed border-border/70 px-4 text-sm font-semibold text-muted-foreground"
    >
      暂无活动记录
    </div>
    <div v-else id="cal-heatmap" class="max-w-full overflow-hidden text-muted-foreground" ref="heatmapContainer"></div>
    <div
        v-if="!isUnavailable"
        class="mt-3 flex w-full flex-wrap items-center justify-end text-xs font-semibold text-muted-foreground"
    >
      <span>少</span>
      <div id="ex-ghDay-legend" class="mx-2 flex"></div>
      <span>多</span>
    </div>
  </div>
</template>

<style scoped>
:deep(.ch-domain-text),
:deep(.ch-subdomain-text) {
  fill: rgb(var(--color-muted-foreground));
}
</style>
