import { Button, COLORS, HeadingLarge, ParagraphMedium } from "@nilfoundation/ui-kit";
import failedIcon from "../../public/icons/confused-face.svg";
import { Box, Icon, Logo } from "../features/components/shared";

interface ErrorScreenProps {
  onRetry: () => void;
}

export const ErrorScreen = ({ onRetry }: ErrorScreenProps) => {
  const handleContactSupport = async () => {
    const endpointUrl = import.meta.env.VITE_APP_COMMUNITY;
    if (endpointUrl) {
      await chrome.tabs.create({ url: endpointUrl });
    } else {
      console.error("Environment variable VITE_APP_COMMUNITY is not set.");
    }
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
        <Icon src={failedIcon} alt="Error Icon" size={124} iconSize="100%" />

        {/* Heading */}
        <HeadingLarge>Something went wrong</HeadingLarge>

        <ParagraphMedium $style={{ color: COLORS.gray300 }}>
          We're working on fixing the issue. Please try again or contact us
        </ParagraphMedium>
      </Box>

      {/* Bottom: Buttons */}
      <Box $align="center" $gap="8px" $style={{ width: "100%" }}>
        <Button
          onClick={onRetry}
          overrides={{
            Root: {
              style: {
                width: "100%",
                height: "48px",
              },
            },
          }}
        >
          Try Again
        </Button>

        {/* Get Endpoint Button */}
        <Button
          onClick={handleContactSupport}
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
          Contact support
        </Button>
      </Box>
    </Box>
  );
};
