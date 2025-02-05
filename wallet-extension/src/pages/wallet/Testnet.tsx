import { Button, COLORS, HeadingLarge, ParagraphMedium } from "@nilfoundation/ui-kit";
import { useTranslation } from "react-i18next";
import outdoorIcon from "../../../public/icons/outdoor.svg";
import { Box, Icon } from "../../features/components/shared";
import { ScreenHeader } from "../../features/components/shared";
import { WalletRoutes } from "../../router";

export const Testnet = () => {
  const { t } = useTranslation("translation");
  const handleCommunityButton = async () => {
    const endpointUrl = import.meta.env.VITE_APP_COMMUNITY;
    if (endpointUrl) {
      await chrome.tabs.create({ url: endpointUrl });
    } else {
      console.error("Environment variable VITE_APP_COMMUNITY is not set.");
    }
  };

  return (
    <Box $align="center" $justify="space-between" $padding="24px" $style={{ height: "100vh" }}>
      {/* Header */}
      <ScreenHeader route={WalletRoutes.WALLET.BASE} title="" />

      {/* Center: Icon, Heading, and Paragraph */}
      <Box $align="center" $gap="16px" $padding="0" $style={{ textAlign: "center" }}>
        <Icon
          src={outdoorIcon}
          alt="Testnet Mode"
          size={124}
          iconSize="100%"
          hoverBackground="transparent"
        />
        <HeadingLarge>{t("wallet.testNet.title")}</HeadingLarge>
        <ParagraphMedium $style={{ color: COLORS.gray300 }}>
          {t("wallet.testNet.text")}
        </ParagraphMedium>
      </Box>

      {/* Bottom: Button */}
      <Box $align="center" $padding="0" $style={{ width: "100%" }}>
        <Button
          onClick={handleCommunityButton}
          overrides={{
            Root: {
              style: {
                width: "100%",
                height: "48px",
              },
            },
          }}
        >
          {t("wallet.testNet.button")}
        </Button>
      </Box>
    </Box>
  );
};
