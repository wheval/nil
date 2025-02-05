import { Button, COLORS, HeadingMedium, Input, ParagraphSmall } from "@nilfoundation/ui-kit";
import { useStore } from "effector-react";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import CheckmarkIcon from "../../../public/icons/checkmark.svg";
import pathIcon from "../../../public/icons/path.svg";
import walletIcon from "../../../public/icons/wallet.svg";
import { Box, Icon, ScreenHeader } from "../../features/components/shared";
import { $smartAccount } from "../../features/store/model/smartAccount.ts";
import { WalletRoutes } from "../../router";

export const Receive = () => {
  const { t } = useTranslation("translation");
  const smartAccount = useStore($smartAccount);

  // State for button loading and text
  const [isLoading, setIsLoading] = useState(false);
  const [showCheckmark, setShowCheckmark] = useState(false);
  const [buttonText, setButtonText] = useState(t("wallet.receivePage.copyButton"));

  // Copy smartAccount address to clipboard
  const handleCopyToClipboard = async () => {
    if (smartAccount?.address) {
      setIsLoading(true); // Start loading state
      try {
        await navigator.clipboard.writeText(smartAccount.address);
        console.log("SmartAccount address copied to clipboard:", smartAccount.address);

        // Update button text and show checkmark
        setButtonText(t("wallet.receivePage.copiedButton"));
        setShowCheckmark(true);

        // Reset button text and hide checkmark after 10 seconds
        setTimeout(() => {
          setButtonText(t("wallet.receivePage.copyButton"));
          setShowCheckmark(false);
        }, 10000);
      } catch (error) {
        console.error("Error copying smartAccount address:", error);
      } finally {
        setIsLoading(false); // End loading state
      }
    }
  };

  return (
    <Box
      $align="stretch"
      $justify="flex-start"
      $padding="24px"
      $style={{ height: "100vh", boxSizing: "border-box" }}
    >
      {/* Header */}
      <ScreenHeader route={WalletRoutes.WALLET.BASE} title={t("wallet.receivePage.title")} />

      {/* Centered Content */}
      <Box $align="center" $gap="16px" $padding="0" $style={{ marginTop: "24px" }}>
        <Icon src={walletIcon} alt="Wallet Icon" size={64} iconSize="100%" />
        <HeadingMedium $style={{ color: COLORS.gray50 }}>
          {t("wallet.receivePage.mainWallet")}
        </HeadingMedium>
      </Box>

      {/* Input and Copy Button */}
      <Box
        $align="center"
        $gap="8px"
        $padding="0"
        $style={{
          flexDirection: "row",
          width: "100%",
          paddingTop: "25px",
        }}
      >
        <Input
          id="walletAddress"
          value={smartAccount?.address || ""}
          readOnly
          overrides={{
            Root: {
              style: {
                flex: "1",
                whiteSpace: "nowrap",
                overflow: "hidden",
                textOverflow: "ellipsis",
              },
            },
          }}
        />
        <Button
          onClick={handleCopyToClipboard}
          isLoading={isLoading}
          overrides={{
            Root: {
              style: {
                width: "120px",
                padding: "0 12px",
                height: "48px",
                backgroundColor: COLORS.gray50,
                color: COLORS.gray800,
                ":hover": {
                  backgroundColor: COLORS.gray100,
                },
              },
            },
          }}
        >
          {showCheckmark ? (
            <Box $align="center" $justify="center" $style={{ flexDirection: "row", gap: "8px" }}>
              <Icon
                src={CheckmarkIcon}
                size={16}
                color={COLORS.gray200}
                iconSize="100%"
                background={"transparent"}
              />
              {buttonText}
            </Box>
          ) : (
            buttonText
          )}
        </Button>
      </Box>

      {/* Activity Section */}
      <Box $align="center" $padding="40px 0" $style={{ textAlign: "center" }}>
        <Box
          $align="center"
          $justify="center"
          $style={{ width: "56px", height: "56px", margin: "0 auto" }}
        >
          <Icon src={pathIcon} alt="No Activity" size={56} iconSize="100%" />
        </Box>
        <ParagraphSmall $style={{ marginTop: "16px", color: COLORS.gray50 }}>
          {t("wallet.receivePage.noActivityText")}
        </ParagraphSmall>
      </Box>
    </Box>
  );
};
