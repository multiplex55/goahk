export type WindowItem = {
  id: string;
  title: string;
  processName: string;
};

export type PropertyItem = {
  name: string;
  value: string;
};

export type PatternAction = {
  id: string;
  label: string;
};

export type TreeNode = {
  id: string;
  name: string;
  hasChildren: boolean;
};
