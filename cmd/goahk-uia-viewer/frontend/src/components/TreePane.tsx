import { TreeNode } from '../types';

type TreePaneProps = {
  nodesByID: Record<string, TreeNode>;
  childrenByParentID: Record<string, string[]>;
  expandedByID: Record<string, boolean>;
  onToggleNode: (id: string) => void;
  selectedNodeId?: string;
  onSelectNode?: (id: string) => void;
};

const ROOT_PARENT_ID = '__root__';

function TreeBranch({ parentID, depth, nodesByID, childrenByParentID, expandedByID, onToggleNode, selectedNodeId, onSelectNode }: {
  parentID: string;
  depth: number;
  nodesByID: Record<string, TreeNode>;
  childrenByParentID: Record<string, string[]>;
  expandedByID: Record<string, boolean>;
  onToggleNode: (id: string) => void;
  selectedNodeId?: string;
  onSelectNode?: (id: string) => void;
}) {
  const childIDs = childrenByParentID[parentID] ?? [];
  return (
    <ul className="tree-list">
      {childIDs.map((childID) => {
        const node = nodesByID[childID];
        if (!node) {
          return null;
        }

        const isExpanded = !!expandedByID[childID];
        const hasChildren = node.hasChildren;

        return (
          <li key={childID}>
            <div className="tree-row" style={{ paddingLeft: `${depth * 16}px` }}>
              <button
                type="button"
                aria-label={hasChildren ? `${isExpanded ? 'Collapse' : 'Expand'} ${node.name}` : `Leaf ${node.name}`}
                onClick={() => {
                  if (hasChildren) {
                    onToggleNode(childID);
                  }
                }}
              >
                {hasChildren ? (isExpanded ? '▾' : '▸') : '•'}
              </button>
              <button
                type="button"
                aria-label={`Select ${node.name}`}
                className={selectedNodeId === childID ? 'selected' : ''}
                onClick={() => onSelectNode?.(childID)}
              >
                {node.name}
              </button>
            </div>
            {hasChildren && isExpanded ? (
              <TreeBranch
                parentID={childID}
                depth={depth + 1}
                nodesByID={nodesByID}
                childrenByParentID={childrenByParentID}
                expandedByID={expandedByID}
                onToggleNode={onToggleNode}
                selectedNodeId={selectedNodeId}
                onSelectNode={onSelectNode}
              />
            ) : null}
          </li>
        );
      })}
    </ul>
  );
}

export default function TreePane({ nodesByID, childrenByParentID, expandedByID, onToggleNode, selectedNodeId, onSelectNode }: TreePaneProps) {
  return (
    <section className="pane" aria-label="uia tree pane">
      <h2>UIA Tree (Lazy)</h2>
      <TreeBranch
        parentID={ROOT_PARENT_ID}
        depth={0}
        nodesByID={nodesByID}
        childrenByParentID={childrenByParentID}
        expandedByID={expandedByID}
        onToggleNode={onToggleNode}
        selectedNodeId={selectedNodeId}
        onSelectNode={onSelectNode}
      />
    </section>
  );
}
