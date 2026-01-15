// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {AccessControl} from "@openzeppelin/contracts/access/AccessControl.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

contract OwnaFarmVault is AccessControl, ReentrancyGuard {
    using SafeERC20 for IERC20;
    
    error OnlyFarmNFT();
    error InsufficientReserve();
    error FarmNFTAlreadySet();
    
    bytes32 public constant ADMIN_ROLE = keccak256("ADMIN_ROLE");
    
    IERC20 public immutable GOLD;
    address public farmNFT;
    uint256 public totalYieldReserve;
    
    event YieldDeposited(uint256 amount);
    event YieldWithdrawn(address indexed to, uint256 amount);
    event FarmNFTSet(address indexed newFarmNFT);
    
    constructor(address gold_) {
        GOLD = IERC20(gold_);
        _grantRole(DEFAULT_ADMIN_ROLE, msg.sender);
        _grantRole(ADMIN_ROLE, msg.sender);
    }
    
    function setFarmNFT(address newFarmNFT) external onlyRole(ADMIN_ROLE) {
        if (farmNFT != address(0)) revert FarmNFTAlreadySet();
        farmNFT = newFarmNFT;
        emit FarmNFTSet(newFarmNFT);
    }
    
    function depositYield(uint256 amount) external onlyRole(ADMIN_ROLE) {
        GOLD.safeTransferFrom(msg.sender, address(this), amount);
        unchecked { totalYieldReserve += amount; }
        emit YieldDeposited(amount);
    }
    
    function withdrawYield(address to, uint256 amount) external nonReentrant {
        if (msg.sender != farmNFT) revert OnlyFarmNFT();
        if (amount > totalYieldReserve) revert InsufficientReserve();
        unchecked { totalYieldReserve -= amount; }
        GOLD.safeTransfer(to, amount);
        emit YieldWithdrawn(to, amount);
    }
    
    function emergencyWithdraw(address token, uint256 amount) external onlyRole(DEFAULT_ADMIN_ROLE) {
        IERC20(token).safeTransfer(msg.sender, amount);
    }
}