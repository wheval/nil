import { COLORS, HeadingMedium } from "@nilfoundation/ui-kit";
import { useNavigate } from "react-router-dom";
import backArrow from "../../../../public/icons/arrows/arrow-back.svg";
import { Box, Icon } from "./index.ts";

interface ScreenHeaderProps {
  route: string;
  title: string;
}

export const ScreenHeader: React.FC<ScreenHeaderProps> = ({ route, title }) => {
  const navigate = useNavigate();

  const handleBack = () => {
    navigate(route);
  };

  return (
    <Box
      $align="center"
      $gap="12px"
      $style={{
        flexDirection: "row",
        width: "100%",
      }}
    >
      {/* Back Icon */}
      <Icon
        src={backArrow}
        alt="Back"
        size={32}
        background={COLORS.gray800}
        hoverBackground={COLORS.gray700}
        round={false}
        pointer={true}
        onClick={handleBack}
      />

      {/* Back Text */}
      <HeadingMedium $style={{ color: COLORS.gray50 }}>{title}</HeadingMedium>
    </Box>
  );
};
