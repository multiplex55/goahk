export type WindowItem = {
  id: string;
  title: string;
  processName: string;
};

export type PropertyItem = {
  name: string;
  value: string;
};

export type Rect = {
  left: number;
  top: number;
  width: number;
  height: number;
};

export type Selector = {
  automationId?: string;
  name?: string;
  controlType?: string;
  className?: string;
  frameworkId?: string;
};

export type SelectorCandidate = {
  rank: number;
  selector: Selector;
  rationale?: string;
  score?: number;
  source?: string;
};

export type WindowInfoDetails = {
  title?: string;
  hwnd?: string;
  text?: string;
  rect?: Rect;
  class?: string;
  process?: string;
  pid?: number;
};

export type ElementDetails = {
  controlType?: string;
  localizedControlType?: string;
  name?: string;
  value?: string;
  automationId?: string;
  bounds?: Rect;
  helpText?: string;
  accessKey?: string;
  acceleratorKey?: string;
  isKeyboardFocusable: boolean;
  hasKeyboardFocus: boolean;
  itemType?: string;
  itemStatus?: string;
  isEnabled: boolean;
  isPassword: boolean;
  isOffscreen: boolean;
  frameworkId?: string;
  isRequiredForForm: boolean;
  status?: string;
};

export type PatternAction = {
  id: string;
  label: string;
  supported?: boolean;
  requiresInput?: boolean;
};

export type TreeNode = {
  id: string;
  name: string;
  hasChildren: boolean;
};

export type NodeDetailsView = {
  windowInfo?: WindowInfoDetails;
  element?: ElementDetails;
  bestSelector?: string;
  selectorPath?: {
    bestSelector?: Selector;
    fullPath?: { nodeID: string; name?: string }[];
    selectorSuggestions?: SelectorCandidate[];
  };
};
