import { getShardIdFromAddress } from "@nilfoundation/niljs";
import { Button, COLORS, CopyButton, ParagraphMedium, ParagraphSmall } from "@nilfoundation/ui-kit";
import { StatefulTooltip } from "baseui/tooltip";
import { useStore } from "effector-react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import linkIcon from "../../../../../public/icons/link.svg";
import settingsIcon from "../../../../../public/icons/settings.svg";
import walletIcon from "../../../../../public/icons/wallet.svg";
import { WalletRoutes } from "../../../../router";
import { $smartAccount } from "../../../store/model/smartAccount.ts";
import { formatAddress } from "../../../utils";
import { Box, Icon } from "../../shared";

export const Header = () => {
  const { t } = useTranslation("translation");
  const navigate = useNavigate();
  const smartAccount = useStore($smartAccount);

  const handleSettings = () => {
    navigate(WalletRoutes.WALLET.SETTINGS);
  };

  const handleNavigate = async () => {
    const endpointUrl = import.meta.env.VITE_NIL_EXPLORER;
    if (endpointUrl && smartAccount?.address) {
      const urlWithAddress = `${endpointUrl}address/${smartAccount.address}`;
      await chrome.tabs.create({ url: urlWithAddress });
    } else {
      console.error("Environment variable VITE_NIL_EXPLORER or smartAccount address is not set.");
    }
  };

  return (
    <Box
      $align="center"
      $justify="space-between"
      $style={{
        flexDirection: "row",
        width: "100%",
      }}
    >
      {/* Left Section */}
      <Box $align="center" $gap="8px" $style={{ flexDirection: "row" }}>
        <Icon src={walletIcon} alt="Wallet Icon" size={64} iconSize="100%" />
        <Box $align="flex-start">
          <ParagraphMedium $style={{ color: COLORS.gray50 }}>
            {t("wallet.header.mainWallet")}
          </ParagraphMedium>
          <Box $align="center" $gap="4px" $style={{ flexDirection: "row" }}>
            <StatefulTooltip
              content={() =>
                `Shard ID: ${smartAccount ? getShardIdFromAddress(smartAccount.address) : ""}`
              }
              showArrow={true}
              placement="bottom"
              overrides={{
                Inner: {
                  style: {
                    backgroundColor: COLORS.gray50,
                  },
                },
              }}
            >
              <ParagraphSmall $style={{ color: COLORS.gray200 }}>
                {smartAccount ? formatAddress(smartAccount.address) : ""}
              </ParagraphSmall>
            </StatefulTooltip>

            {/* Add Icons here */}
            <CopyButton textToCopy={smartAccount?.address} />

            <StatefulTooltip
              content={() => "Open in Explorer"}
              showArrow={true}
              placement="bottom"
              overrides={{
                Inner: {
                  style: {
                    backgroundColor: COLORS.gray50,
                  },
                },
              }}
            >
              <div
                style={{
                  display: "inline-flex",
                  alignItems: "center",
                  justifyContent: "center",
                  padding: "4px",
                }}
              >
                <Icon
                  pointer={true}
                  src={linkIcon}
                  alt="Link"
                  size={20}
                  iconSize="100%"
                  background={"transparent"}
                  onClick={handleNavigate}
                />
              </div>
            </StatefulTooltip>
          </Box>
        </Box>
      </Box>

      {/* Right Section */}
      <Box
        $justify="flex-end"
        $align="center"
        $gap="8px"
        $style={{ flexDirection: "row", flex: "1" }}
      >
        <Button
          onClick={() => {
            navigate(WalletRoutes.WALLET.TESTNET);
          }}
          overrides={{
            Root: {
              style: {
                backgroundColor: COLORS.yellow800,
                color: COLORS.yellow200,
                width: "65px",
                height: "30px",
                fontSize: "12px",
                padding: "0",
                borderRadius: "4px",
                ":hover": {
                  backgroundColor: COLORS.yellow700,
                },
              },
            },
          }}
        >
          {t("wallet.header.testnetButton")}
        </Button>
        <Icon
          src={settingsIcon}
          alt="Settings Icon"
          size={32}
          background={COLORS.gray800}
          hoverBackground={COLORS.gray700}
          round={false}
          pointer={true}
          onClick={handleSettings}
        />
      </Box>
    </Box>
  );
};
