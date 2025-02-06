import { Button, COLORS, HeadingLarge, ParagraphMedium } from "@nilfoundation/ui-kit";
import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import failedIcon from "../../../public/icons/confused-face.svg";
import { clearState } from "../../background/storage";
import { Box, Icon, Logo } from "../../features/components/shared";
import { resetEndpoint } from "../../features/store/model/endpoint.ts";
import { resetGlobalError } from "../../features/store/model/error.ts";
import {
  resetSmartAccount,
  retrySmartAccountCreation,
} from "../../features/store/model/smartAccount.ts";
import { WalletRoutes } from "../../router";

export const ErrorPage = () => {
  const navigate = useNavigate();
  const { t } = useTranslation("translation");

  // Clears state on page load
  useEffect(() => {
    const clearStateOnPageLoad = async () => {
      try {
        await clearState();
        console.log("State cleared successfully on page load.");
      } catch (error) {
        console.error("Error clearing state on page load:", error);
      }
    };

    clearStateOnPageLoad();
  }, []);

  // Handles try again button press
  const handleTryAgain = () => {
    resetGlobalError();
    resetSmartAccount();
    retrySmartAccountCreation();
    navigate(WalletRoutes.GET_STARTED.LOADING);
  };

  // Handles Go Back button
  const handleGoBack = async () => {
    resetGlobalError();
    resetSmartAccount();
    resetEndpoint();
    navigate(WalletRoutes.GET_STARTED.SET_ENDPOINT);
  };

  return (
    <Box $align="center" $justify="space-between" $padding="24px" $style={{ height: "100vh" }}>
      {/* Top: Logo */}
      <Box>
        <Logo size={40} />
      </Box>

      {/* Center: Icon, Heading, and Paragraph */}
      <Box $align="center" $gap="16px" $padding="0" $style={{ textAlign: "center", width: "100%" }}>
        {/* Icon */}
        <Icon src={failedIcon} alt="Asterisk Icon" size={124} iconSize="100%" />

        {/* Heading */}
        <HeadingLarge>{t("getStarted.error.heading")}</HeadingLarge>

        <ParagraphMedium $style={{ color: COLORS.gray300 }}>
          {t("getStarted.error.description")}
        </ParagraphMedium>
      </Box>

      {/* Bottom: Buttons */}
      <Box $align="center" $gap="8px" $style={{ width: "100%" }}>
        {/* Connect Button */}
        <Button
          onClick={handleTryAgain}
          overrides={{
            Root: {
              style: {
                width: "100%",
                height: "48px",
              },
            },
          }}
        >
          {t("getStarted.error.tryAgain")}
        </Button>

        {/* Get Endpoint Button */}
        <Button
          onClick={handleGoBack}
          overrides={{
            Root: {
              style: {
                width: "100%",
                height: "48px",
                backgroundColor: COLORS.gray800,
                color: COLORS.gray200,
                ":hover": {
                  backgroundColor: COLORS.gray700,
                },
              },
            },
          }}
        >
          {t("getStarted.error.goBack")}
        </Button>
      </Box>
    </Box>
  );
};
