import { TreeNode } from '../types';

type TreePaneProps = {
  rootNodes: TreeNode[];
  expandedNodeIds: Set<string>;
  onToggleNode: (id: string) => void;
  selectedNodeId?: string;
  onSelectNode?: (id: string) => void;
};

export default function TreePane({ rootNodes, expandedNodeIds, onToggleNode, selectedNodeId, onSelectNode }: TreePaneProps) {
  return (
    <section className="pane" aria-label="uia tree pane">
      <h2>UIA Tree (Lazy)</h2>
      <ul className="tree-list">
        {rootNodes.map((node) => (
          <li key={node.id}>
            <button
              type="button"
              className={selectedNodeId === node.id ? 'selected' : ''}
              onClick={() => {
                if (node.hasChildren) {
                  onToggleNode(node.id);
                }
                onSelectNode?.(node.id);
              }}
            >
              {node.hasChildren ? (expandedNodeIds.has(node.id) ? '▾' : '▸') : '•'} {node.name}
            </button>
          </li>
        ))}
      </ul>
    </section>
  );
}
