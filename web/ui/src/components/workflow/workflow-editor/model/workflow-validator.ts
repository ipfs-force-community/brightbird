import { Cell, CellView, Graph, JQuery, Node, Point } from '@antv/x6';
import { CustomX6NodeProxy } from './data/custom-x6-node-proxy';
import nodeWarningIcon from '../svgs/node-warning.svg';
import { IWorkflow } from './data/common';

export type ClickNodeWarningCallbackFnType = (nodeId: string) => void;

function isWarning(node: Node): boolean {
  return node.hasTool('button');
}

export class WorkflowValidator {
  private readonly graph: Graph;
  private readonly proxy: any;
  private readonly workflowData: IWorkflow;

  constructor(graph: Graph, proxy: any, workflowData: IWorkflow) {
    this.graph = graph;
    this.proxy = proxy;
    this.workflowData = workflowData;
  }

  addWarning(node: Node, clickNodeWarningCallback: ClickNodeWarningCallbackFnType): void {
    if (isWarning(node)) {
      return;
    }

    node.addTools({
      name: 'button',
      args: {
        markup: [
          {
            tagName: 'image',
            attrs: {
              width: 24,
              height: 24,
              'xlink:href': nodeWarningIcon,
              cursor: 'pointer',
            },
          },
        ],
        x: '100%',
        y: 0,
        offset: { x: -13, y: -11 },
        onClick: ({ cell: { id } }: { e: JQuery.MouseDownEvent; cell: Cell; view: CellView }) =>
          clickNodeWarningCallback(id),
      },
    });
  }

  removeWarning(node: Node): void {
    if (!isWarning(node)) {
      return;
    }

    node.removeTool('button');
  }

  /**
   * 校验所有节点
   * @throws 尚未通过校验时，抛异常
   */
  async checkNodes(): Promise<void> {
    const nodes = this.graph.getNodes();

    if (nodes.length === 0) {
      throw new Error('未存在任何节点');
    }

    if (nodes.length > 1) {
      const nodeSet = new Set<Node>();
      this.graph.getEdges().forEach(edge => {
        nodeSet.add(edge.getSourceNode()!);
        nodeSet.add(edge.getTargetNode()!);
      });
      const connectedNodes = Array.from(nodeSet);
      const unconnectedNodes = nodes.filter(node => !connectedNodes.includes(node));
      if (unconnectedNodes.length > 0) {
        const nodeName = new CustomX6NodeProxy(unconnectedNodes[0]).getData().getName();
        // 存在未连接的节点
        throw new Error(`${nodeName}节点：尚未连接任何其他节点`);
      }
    }

    const workflowNodes = nodes.map(node => new CustomX6NodeProxy(node).getData(this.graph, this.workflowData));
    for (const workflowNode of workflowNodes) {
      try {
        await workflowNode.validate();
      } catch ({ errors }) {
        throw new Error(`${workflowNode.getName()}节点：${errors[0].message}`);
      }
    }
  }

  checkDroppingNode(node: Node, mousePosition: Point.PointLike, nodePanelRect: DOMRect): boolean {
    if (!this.checkDroppingPosition(mousePosition, nodePanelRect)) {
      return false;
    }

    return true;
  }

  private checkDroppingPosition(mousePosition: Point.PointLike, nodePanelRect: DOMRect): boolean {
    const { x: mousePosX, y: mousePosY } = mousePosition;
    const { x, y, width, height } = nodePanelRect;
    const maxX = x + width;
    const maxY = y + height;

    if (mousePosX >= x && mousePosX <= maxX && mousePosY >= y && mousePosY <= maxY) {
      // 在节点面板中拖放时，失败
      return false;
    }

    return true;
  }

}
