import { Button, COLORS, HeadingLarge, Input } from "@nilfoundation/ui-kit";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import asteriskIcon from "../../../public/icons/asterisk.svg";
import { Box, Icon, InputErrorMessage, Logo } from "../../features/components/shared";
import { setEndpoint } from "../../features/store/model/endpoint";
import { type ValidationResult, validateRpcEndpoint } from "../../features/utils";
import { WalletRoutes } from "../../router";
import { setInitialTokens } from "../../features/store/model/token.ts";

export const SetEndpoint = () => {
  const navigate = useNavigate();
  const { t } = useTranslation("translation");
  const [inputValue, setInputValue] = useState("");
  const [error, setError] = useState("");

  // Handles connect button press
  const handleConnect = () => {
    const rpcValidation: ValidationResult = validateRpcEndpoint(inputValue);
    if (!rpcValidation.isValid) {
      setError(rpcValidation.error);
      return;
    }

    setError("");
    setEndpoint(inputValue);

    navigate(WalletRoutes.GET_STARTED.LOADING);
  };

  // Opens a new browser tab to fetch the endpoint URL if available
  const handleGetEndpoint = async () => {
    const endpointUrl = import.meta.env.VITE_GET_ENDPOINT_URL;
    if (endpointUrl) {
      await chrome.tabs.create({ url: endpointUrl });
    } else {
      console.error("Environment variable REACT_APP_GET_ENDPOINT_URL is not set.");
    }
  };

  // Whenever the user types, clear the error
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setError("");
    setInputValue(e.target.value);
  };

  return (
    <Box $align="center" $justify="space-between" $padding="24px" $style={{ height: "100vh" }}>
      {/* Top: Logo */}
      <Box>
        <Logo size={40} />
      </Box>

      {/* Center: Icon, Heading, and Input */}
      <Box $align="center" $gap="16px" $padding="0" $style={{ textAlign: "center", width: "100%" }}>
        {/* Icon */}
        <Icon src={asteriskIcon} alt="Asterisk Icon" size={124} iconSize="100%" />

        {/* Heading */}
        <HeadingLarge>{t("getStarted.setEndpoint.heading")}</HeadingLarge>

        {/* The input for endpoint */}
        <Input
          error={error !== ""}
          placeholder={t("getStarted.setEndpoint.inputPlaceholder")}
          value={inputValue}
          onChange={handleInputChange}
          overrides={{
            Root: {
              style: {
                width: "100%",
                height: "48px",
              },
            },
          }}
        />

        {/* If error is not empty, display it */}
        <InputErrorMessage error={error} />
      </Box>

      {/* Bottom: Buttons */}
      <Box $align="center" $gap="8px" $style={{ width: "100%" }}>
        {/* Connect Button */}
        <Button
          onClick={handleConnect}
          overrides={{
            Root: {
              style: {
                width: "100%",
                height: "48px",
              },
            },
          }}
        >
          {t("getStarted.setEndpoint.connectButton")}
        </Button>

        {/* Get Endpoint Button */}
        <Button
          onClick={handleGetEndpoint}
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
          {t("getStarted.setEndpoint.getEndpointButton")}
        </Button>
      </Box>
    </Box>
  );
};
