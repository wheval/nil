import type { Hex } from "@nilfoundation/niljs";
import type { SmartAccountV1 } from "@nilfoundation/niljs";
import {
  Button,
  COLORS,
  HeadingMedium,
  Input,
  ParagraphSmall,
  ParagraphXSmall,
} from "@nilfoundation/ui-kit";
import { useStore } from "effector-react";
import { useCallback, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { parseEther } from "viem";
import { sendToken } from "../../features/blockchain";
import {
  Box,
  InputErrorMessage,
  MainAddressInput,
  ScreenHeader,
  TokenInput,
} from "../../features/components/shared";
import { TokenNames } from "../../features/components/token";
import { $smartAccount } from "../../features/store/model/smartAccount.ts";
import {
  $balance,
  $balanceToken,
  $tokens,
  getBalanceForToken,
} from "../../features/store/model/token.ts";
import {
  convertWeiToEth,
  fetchEstimatedFee,
  getTokens,
  validateSendAmount,
} from "../../features/utils";
import { validateSmartAccountAddress } from "../../features/utils/inputValidation";
import { WalletRoutes } from "../../router";

export const Send = () => {
  const { t } = useTranslation("translation");
  const navigate = useNavigate();
  const smartAccount = useStore($smartAccount);
  const availableTokens = useStore($tokens);
  const balanceTokens = useStore($balanceToken);
  const nilBalance = useStore($balance);

  const tokens = getTokens(availableTokens, true);

  const [toAddress, setToAddress] = useState("");
  const [amount, setAmount] = useState("");
  const [addressError, setAddressError] = useState("");
  const [amountError, setAmountError] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [estimatedFee, setEstimatedFee] = useState("");
  const [selectedToken, setSelectedToken] = useState(tokens[0]);

  const balance = getBalanceForToken(selectedToken.address, nilBalance ?? 0n, balanceTokens ?? {});

  // Validation Function
  const validateTransaction = useCallback(() => {
    if (!smartAccount || !balanceTokens) return false;

    const addressValidation = validateSmartAccountAddress(toAddress, smartAccount.address);
    const amountValidation = validateSendAmount(amount, selectedToken.label, balance);

    setAddressError(addressValidation.isValid ? "" : addressValidation.error);
    setAmountError(amountValidation ?? "");

    return addressValidation.isValid && !amountValidation;
  }, [smartAccount, balanceTokens, toAddress, amount, selectedToken, balance]);

  // Debounced Gas Fee Calculation
  useEffect(() => {
    if (!toAddress || !amount || !balanceTokens || !balance || !smartAccount) {
      setEstimatedFee("");
      return;
    }

    if (!validateTransaction()) return;

    const timeoutId = setTimeout(async () => {
      try {
        const fee = await fetchEstimatedFee(
          smartAccount as SmartAccountV1,
          toAddress as Hex,
          amount,
          selectedToken.address,
        );
        setEstimatedFee(fee);

        // Adjust amount to leave gas for NIL transactions
        if (selectedToken.label === TokenNames.NIL && parseEther(amount) === balance) {
          const adjustedValue = parseEther(amount) - parseEther(String(Number(fee) * 2));
          setAmount(convertWeiToEth(adjustedValue));
        }
      } catch (err) {
        console.log(err);
        setEstimatedFee("Error estimating fee");
      }
    }, 500); // **Debounce effect - Waits 500ms after last change before running**

    return () => clearTimeout(timeoutId);
  }, [toAddress, amount, selectedToken, balanceTokens, balance, smartAccount, validateTransaction]);

  const handleSend = async () => {
    if (!validateTransaction()) return;

    if (balance === parseEther(amount)) {
      setAmountError("Leave funds for gas fees");
      return;
    }

    if (!smartAccount) return;

    setIsLoading(true);
    try {
      await sendToken({
        smartAccount: smartAccount as SmartAccountV1,
        to: toAddress as Hex,
        value: Number(amount),
        tokenAddress: selectedToken.address,
        symbol: selectedToken.label,
      });
      console.log(`Successfully sent ${amount} ${selectedToken.label} to ${toAddress}`);
      navigate(WalletRoutes.WALLET.BASE);
    } catch (error) {
      console.error("Send failed:", error);
    } finally {
      setIsLoading(false);
      navigate(WalletRoutes.WALLET.BASE);
    }
  };

  return (
    <Box
      $style={{
        display: "flex",
        flexDirection: "column",
        height: "100vh",
        padding: "24px",
        boxSizing: "border-box",
      }}
    >
      {/* Header */}
      <ScreenHeader route={WalletRoutes.WALLET.BASE} title={t("wallet.sendPage.title")} />

      {/* From Section */}
      <Box $style={{ marginTop: "24px" }}>
        <HeadingMedium $style={{ color: COLORS.gray50, marginBottom: "5px" }}>
          {t("wallet.sendPage.fromSection.heading")}
        </HeadingMedium>
        <MainAddressInput />
      </Box>

      {/* To Section */}
      <Box $style={{ marginTop: "12px" }}>
        <HeadingMedium $style={{ color: COLORS.gray50, marginBottom: "8px" }}>
          {t("wallet.sendPage.toSection.heading")}
        </HeadingMedium>
        <Input
          error={addressError}
          placeholder={t("wallet.sendPage.toSection.inputPlaceholder")}
          value={toAddress}
          onChange={(e) => {
            setAddressError("");
            setToAddress(e.target.value);
            setEstimatedFee("");
          }}
        />
        {addressError && <InputErrorMessage error={addressError} style={{ marginTop: "5px" }} />}
      </Box>

      {/* Amount Section */}
      <Box $style={{ marginTop: "12px" }}>
        <HeadingMedium $style={{ color: COLORS.gray50, marginBottom: "5px" }}>
          {t("wallet.sendPage.amountSection.heading")}
        </HeadingMedium>
        <TokenInput
          error={amountError}
          selectedToken={selectedToken}
          tokens={tokens}
          onTokenChange={(params) => {
            setEstimatedFee("");
            const selected = tokens.find((token) => params.value[0]?.label === token.label);
            if (selected) setSelectedToken(selected);
          }}
          value={amount}
          onChange={(e) => {
            setAmount(e.target.value);
            // Debounce clearing estimated fee
            setTimeout(() => {
              setEstimatedFee("");
            }, 500);
          }}
        />
        {amountError && amountError !== "Insufficient Funds" && (
          <InputErrorMessage error={amountError} style={{ marginTop: "5px" }} />
        )}

        {/* Display balance and "Send Max" button */}
        <Box
          $justify="flex-start"
          $align="center"
          $gap="8px"
          $width="auto"
          $style={{ flexDirection: "row", marginTop: "12px" }}
        >
          <ParagraphSmall
            color={amountError === "Insufficient Funds" ? COLORS.red300 : COLORS.gray200}
          >
            {amountError === "Insufficient Funds" ? `${amountError} - ` : ""}Balance:{" "}
            {selectedToken.label === TokenNames.NIL ? convertWeiToEth(balance) : balance.toString()}
          </ParagraphSmall>

          <Box
            $align="center"
            $justify="center"
            onClick={() => {
              setAmountError("");
              if (selectedToken.label === TokenNames.NIL) {
                setAmount(convertWeiToEth(balance));
                return;
              }
              setAmount(balance.toString());
            }}
            $style={{
              width: "80px", // Slightly wider for better readability
              height: "32px",
              backgroundColor: COLORS.gray800,
              color: COLORS.gray200,
              borderRadius: "8px",
              transition: "all 0.2s ease",
              cursor: "pointer",
              fontWeight: "500",
              ":hover": {
                backgroundColor: COLORS.gray700,
              },
              ":active": {
                backgroundColor: COLORS.gray600,
                transform: "scale(0.98)",
              },
            }}
          >
            <ParagraphXSmall> {t("wallet.sendPage.amountSection.max")}</ParagraphXSmall>
          </Box>
        </Box>
      </Box>

      {/* Estimated Gas Fee */}
      {estimatedFee && !amountError && !addressError && (
        <Box
          $justify="space-between"
          $align="center"
          $width="100%"
          $style={{ flexDirection: "row", marginTop: "30px" }}
        >
          <ParagraphSmall color={COLORS.gray200}>
            {t("wallet.sendPage.amountSection.estimate")}
          </ParagraphSmall>
          <ParagraphSmall
            color={estimatedFee?.toString().includes("Error") ? COLORS.red300 : COLORS.gray200}
          >
            {estimatedFee?.toString().includes("Error") ? estimatedFee : `${estimatedFee} NIL`}
          </ParagraphSmall>
        </Box>
      )}

      {/* Button Section */}
      <Box $style={{ marginTop: "auto", width: "100%" }}>
        <Button
          onClick={handleSend}
          isLoading={isLoading}
          overrides={{ Root: { style: { width: "100%", height: "48px" } } }}
        >
          {t("wallet.sendPage.amountSection.buttonLabel")}
        </Button>
      </Box>
    </Box>
  );
};
