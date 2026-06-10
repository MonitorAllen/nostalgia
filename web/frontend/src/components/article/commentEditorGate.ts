export interface CommentEditorGateState {
  isAuthenticated: boolean
  isActivated: boolean
}

export const shouldRenderCommentEditor = ({
  isAuthenticated,
  isActivated
}: CommentEditorGateState) => isAuthenticated && isActivated
