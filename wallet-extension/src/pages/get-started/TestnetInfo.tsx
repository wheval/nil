import { Button, COLORS, HeadingLarge, ParagraphMedium } from "@nilfoundation/ui-kit";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import testnetMode from "../../../public/icons/testnet-mode.svg";
import { Box, Icon, Logo } from "../../features/components/shared";
import { WalletRoutes } from "../../router";

export const TestnetInfo = () => {
  const navigate = useNavigate();
  const { t } = useTranslation("translation");

  const handleButtonClick = () => {
    navigate(WalletRoutes.GET_STARTED.SET_ENDPOINT);
  };

  return (
    <Box $align="center" $justify="space-between" $padding="24px" $style={{ height: "100vh" }}>
      {/* Top: Logo */}
      <Box $align="center">
        <Logo size={40} />
      </Box>

      {/* Center: Icon, Heading, and Paragraph */}
      <Box $align="center" $gap="16px" $padding="0" $style={{ textAlign: "center" }}>
        <Icon
          src={testnetMode}
          alt="Testnet Mode"
          size={124}
          iconSize="100%"
          hoverBackground="transparent"
        />
        <HeadingLarge>{t("getStarted.testnetInfo.heading")}</HeadingLarge>
        <ParagraphMedium $style={{ color: COLORS.gray300 }}>
          {t("getStarted.testnetInfo.description")}
        </ParagraphMedium>
      </Box>

      {/* Bottom: Button */}
      <Box $align="center" $padding="0" $style={{ width: "100%" }}>
        <Button
          onClick={handleButtonClick}
          overrides={{
            Root: {
              style: {
                width: "100%",
                height: "48px",
              },
            },
          }}
        >
          {t("getStarted.testnetInfo.button")}
        </Button>
      </Box>
    </Box>
  );
};
