declare module 'jsoneditor' {
  export interface JSONEditorEventNode {
    field?: string
    path?: Array<string | number>
    value?: unknown
  }

  export interface JSONEditorMenuItem {
    text?: string
    title?: string
    className?: string
    type?: string
    submenu?: JSONEditorMenuItem[]
    click?: () => void
  }

  export interface JSONEditorContextMenuNode {
    type?: 'single' | 'multiple' | 'append'
    path?: Array<string | number>
    paths?: Array<Array<string | number>>
  }

  export interface JSONEditorOptions {
    mode?: 'tree' | 'view' | 'form' | 'code' | 'text' | 'preview'
    mainMenuBar?: boolean
    navigationBar?: boolean
    statusBar?: boolean
    search?: boolean
    onEditable?: (node?: JSONEditorEventNode) => boolean | { field: boolean, value: boolean }
    onEvent?: (node: JSONEditorEventNode, event: Event) => void
    onCreateMenu?: (items: JSONEditorMenuItem[], node: JSONEditorContextMenuNode) => JSONEditorMenuItem[]
  }

  export default class JSONEditor {
    constructor(container: HTMLElement, options?: JSONEditorOptions, json?: unknown)
    set(json: unknown): void
    update(json: unknown): void
    destroy(): void
    expandAll(): void
    collapseAll(): void
    search(text: string): void
    refresh(): void
  }
}
