import { COLORS, ParagraphXSmall } from "@nilfoundation/ui-kit";

export const InputErrorMessage = ({ error, style }) => {
  return (
    <div style={{ height: "15px", width: "100%", textAlign: "start", ...style }}>
      {error && <ParagraphXSmall style={{ color: COLORS.red400 }}>{error}</ParagraphXSmall>}
    </div>
  );
};
