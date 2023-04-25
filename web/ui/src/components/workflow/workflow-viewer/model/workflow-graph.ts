import yaml from 'yaml';
import { BaseGraph } from './base-graph';
import { G6Graph } from './graph/g6';
import { X6Graph } from './graph/x6';
import { INodeDefVo } from '@/api/dto/testflow';
import { GraphDirectionEnum } from './data/enumeration';

export class WorkflowGraph {
  readonly graph: BaseGraph
  private readonly resizeObserver: ResizeObserver;
  readonly visibleDsl: string;

  constructor(dsl: string, nodeInfos: INodeDefVo[], container: HTMLElement, direction: GraphDirectionEnum) {
    const obj = yaml.parse(dsl);
    const { 'raw-data': rawData } = obj;
    if (rawData) {
      delete obj['raw-data'];
      this.visibleDsl = yaml.stringify(obj);
    } else {
      this.visibleDsl = dsl;
    }

    this.graph = rawData ? new X6Graph(dsl, container) :
      new G6Graph(dsl, nodeInfos, container, direction);

    const containerParentEl = container.parentElement!;
    this.resizeObserver = new ResizeObserver(() => {
      const { clientWidth, clientHeight } = containerParentEl;
      this.graph.changeSize(clientWidth, clientHeight);
    });
    // 监控容器大小变化
    this.resizeObserver.observe(containerParentEl);
  }

  destroy() {
    // 销毁监控容器大小变化
    this.resizeObserver.disconnect();

    // 销毁画布
    this.graph.destroy();
  }
}