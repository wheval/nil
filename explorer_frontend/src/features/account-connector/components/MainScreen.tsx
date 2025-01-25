import {
  BUTTON_KIND,
  Button,
  COLORS,
  CopyButton,
  LabelLarge,
  LabelMedium,
  LabelSmall,
  PlusIcon,
} from "@nilfoundation/ui-kit";
import type { ButtonOverrides } from "baseui/button";
import { useUnit } from "effector-react";
import { useStyletron } from "styletron-react";
import { formatEther } from "viem";
import { OverflowEllipsis } from "../../shared";
import { ActiveComponent } from "../ActiveComponent";
import {
  $balance,
  $balanceToken,
  $initializingSmartAccountError,
  $initializingSmartAccountState,
  $smartAccount,
  createSmartAccountFx,
  regenrateAccountEvent,
  setActiveComponent,
  topUpSmartAccountBalanceFx,
} from "../model";
import { EndpointInput } from "./EndpointInput";
import { Token } from "./Token";
import { styles } from "./styles";

const btnOverrides: ButtonOverrides = {
  Root: {
    style: ({ $disabled }) => ({
      backgroundColor: $disabled ? `${COLORS.gray400}!important` : "",
      width: "50%",
    }),
  },
};

const MainScreen = () => {
  const [css] = useStyletron();
  const [isPendingSmartAccountCreation] = useUnit([createSmartAccountFx.pending]);
  const [
    smartAccount,
    balance,
    balanceToken,
    initializingSmartAccountState,
    initializingSmartAccountError,
  ] = useUnit([
    $smartAccount,
    $balance,
    $balanceToken,
    $initializingSmartAccountState,
    $initializingSmartAccountError,
  ]);
  const [isPendingTopUp] = useUnit([topUpSmartAccountBalanceFx.pending]);
  const displayBalance = balance === null ? "-" : formatEther(balance);
  const address = smartAccount ? smartAccount.address : null;

  return (
    <div
      className={css({
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        gap: "24px",
        maxHeight: "531px",
        overflowY: "auto",
        overscrollBehavior: "contain",
        paddingRight: "24px",
      })}
    >
      <div
        className={css({
          width: "100%",
          display: "flex",
          flexDirection: "column",
          position: "sticky",
          alignItems: "center",
          gap: "24px",
          top: 0,
          backgroundColor: COLORS.gray800,
        })}
      >
        <LabelLarge>Smart Account</LabelLarge>
        <div>
          {address !== null && (
            <div
              className={css({
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
                gap: "1ch",
              })}
            >
              <LabelMedium width="250px" color={COLORS.gray200}>
                {isPendingSmartAccountCreation ? (
                  <LabelMedium
                    className={css({
                      textAlign: "center",
                    })}
                  >
                    Creating new smart account
                  </LabelMedium>
                ) : (
                  <OverflowEllipsis>{address}</OverflowEllipsis>
                )}
              </LabelMedium>
              <CopyButton textToCopy={address} disabled={address === null} color={COLORS.gray200} />
            </div>
          )}
          <div
            className={css({
              height: "16px",
            })}
          >
            {(isPendingSmartAccountCreation || initializingSmartAccountError) && (
              <LabelSmall
                color={initializingSmartAccountError ? COLORS.red200 : COLORS.gray400}
                className={css({
                  textAlign: "center",
                })}
              >
                {initializingSmartAccountError
                  ? initializingSmartAccountError
                  : initializingSmartAccountState}
              </LabelSmall>
            )}
          </div>
        </div>
        <EndpointInput />
      </div>
      <ul
        className={css({
          width: "100%",
        })}
      >
        <li className={css(styles.menuItem)}>
          <Token balance={displayBalance} name="NIL" isMain />
        </li>
        {balanceToken !== null &&
          Object.keys(balanceToken).length !== 0 &&
          Object.entries(balanceToken).map(([name, balance]) => (
            <>
              <li key={`divider-${name}`} className={css(styles.divider)} />
              <li key={name} className={css(styles.menuItem)}>
                <Token balance={balance.toString()} name={name} isMain={false} />
              </li>
            </>
          ))}
        <li
          className={css({
            ...styles.menuItem,
            paddingTop: "24px",
            paddingLeft: 0,
            paddingRight: 0,
            position: "sticky",
            bottom: 0,
            backgroundColor: COLORS.gray800,
          })}
        >
          <Button
            kind={BUTTON_KIND.primary}
            onClick={() => setActiveComponent(ActiveComponent.Topup)}
            isLoading={isPendingTopUp}
            disabled={isPendingTopUp || isPendingSmartAccountCreation || !smartAccount}
            overrides={btnOverrides}
            className={css({
              whiteSpace: "nowrap",
            })}
          >
            <PlusIcon size={24} />
            Top up
          </Button>
          <Button
            kind={BUTTON_KIND.toggle}
            onClick={() => regenrateAccountEvent()}
            className={css({
              whiteSpace: "nowrap",
            })}
            disabled={isPendingTopUp || isPendingSmartAccountCreation}
            overrides={{
              Root: {
                style: {
                  width: "50%",
                },
              },
            }}
          >
            Regenerate account
          </Button>
        </li>
      </ul>
    </div>
  );
};

export { MainScreen };
